package window_crc32

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

const (
	crcInitVal = 0xffffffff
)

type crc32Roller struct {
	crc32Tab [256]uint
	rolngTab [256]uint
	window   uint
	oldBuf   []byte
	curBuf   []byte
	bufPos   uint
	crc32    uint
	poly     uint
	threads  uint
}

func newCRC(poly uint, window uint, s Serializer, t uint) *crc32Roller {
	c := &crc32Roller{
		crc32:   crcInitVal,
		window:  window,
		poly:    poly,
		threads: t,
	}

	c.generateTable(poly)

	if window != 0 {
		c.curBuf = make([]byte, window)
		if s == nil {
			c.buildRollingTable(window)
		} else {
			c.buildSerializedRollingTable(window, s)
		}
	}

	return c
}

func (c *crc32Roller) UpdateCRC32(b byte) {
	switch c.window {
	case 0:
		c.updateCRC(b)
	default:
		c.rollingUpdate(b)
	}
}

func (c *crc32Roller) Finish() uint {
	return c.crc32 ^ crcInitVal
}

func (c *crc32Roller) generateTable(poly uint) {
	var i uint
	for i = 0; i < 256; i++ {
		r := i
		for j := 0; j < 8; j++ {
			r = (r >> 1) ^ (poly & ^((r & 1) - 1))
		}
		c.crc32Tab[i] = r
	}
}

func (c *crc32Roller) updateCRC(b byte) {
	c.crc32 = c.crc32Tab[(c.crc32^uint(b))&0xFF] ^ ((c.crc32) >> 8)
}

func (c *crc32Roller) rollingUpdate(b byte) {
	c.updateCRC(b)
	if c.oldBuf != nil {
		c.crc32 ^= c.rolngTab[c.oldBuf[c.bufPos]]
	}

	c.curBuf[c.bufPos] = b
	c.bufPos++
	if c.bufPos == c.window {
		c.oldBuf = c.curBuf
		c.curBuf = make([]byte, c.window)
		c.bufPos = 0
	}
}

func (c *crc32Roller) getCRC32() uint {
	return c.crc32
}

func (c *crc32Roller) setCRC(v uint) {
	c.crc32 = v
}

func (c *crc32Roller) buildRollingTable(window uint) {
	var i, p uint
	y := newCRC(CrcPoly, 0, nil, 0)
	y.updateCRC(0)

	for i = 0; i < window-1; i++ {
		y.updateCRC(0)
	}

	var xTable [256]*crc32Roller
	for p := 0; p < 256; p++ {
		xTable[p] = newCRC(c.poly, 0, nil, 0)
		xTable[p].updateCRC(byte(p))
	}

	sem := semaphore.NewWeighted(int64(c.threads))
	var wg sync.WaitGroup
	wg.Add(256)
	for p = 0; p < 256; p++ {
		_ = sem.Acquire(context.Background(), 1)
		go func(pos uint) {
			defer wg.Done()
			defer sem.Release(1)

			for i = 0; i < window; i++ {
				xTable[pos].updateCRC(0)
			}

			c.rolngTab[pos] = xTable[pos].getCRC32() ^ y.getCRC32()
		}(p)
	}
	wg.Wait()
}

func (c *crc32Roller) buildSerializedRollingTable(window uint, s Serializer) {
	var i, p uint
	y := newCRC(c.poly, 0, nil, 1)
	y.updateCRC(0)

	for i = 0; i < window-1; i++ {
		y.updateCRC(0)
	}

	var xTable [256]*crc32Roller

	start := s.GetSerializedPos(window)
	s.SetPoly(c.poly)
	s.SetWindow(window)
	for p := 0; p < 256; p++ {
		xTable[p] = s.GetXVal(p)
	}

	modPos := s.GetOutputMod()

	// thread pool
	sem := semaphore.NewWeighted(int64(c.threads))
	// need to make sure that all elements of table are filled in before we exit
	var wg sync.WaitGroup
	wg.Add(256)

	for p = 0; p < 256; p++ {
		_ = sem.Acquire(context.Background(), 1)
		go func(pos uint) {
			defer wg.Done()
			defer sem.Release(1)

			for i := start; i < window; i++ {
				if i%modPos == 0 {
					s.SetXVal(i, pos, xTable[pos])
				}
				xTable[pos].updateCRC(0)
			}
			if window%modPos == 0 {
				s.SetXVal(window, pos, xTable[pos])
			}

			c.rolngTab[pos] = xTable[pos].getCRC32() ^ y.getCRC32()
		}(p)
	}

	wg.Wait()
}
