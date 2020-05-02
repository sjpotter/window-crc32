package window_crc32

import (
	"encoding/json"
)

type Serializer interface {
	GetSerializedPos(w uint) uint
	SetWindow(w uint)
	GetOutputMod() uint
	GetXVal(p int) *crc32Roller
	SetXVal(i uint, p uint, c *crc32Roller)
	SetPoly(poly uint)
	WriteOut() ([]byte, error)
	ReadIn([]byte) error
}

type emptySerializer struct {
	poly uint
}

func (e *emptySerializer) GetSerializedPos(uint) uint {
	return 0
}

func (e *emptySerializer) SetWindow(uint) {}

func (e *emptySerializer) GetXVal(p int) *crc32Roller {
	c := newCRC(e.poly, 0, nil, 1)
	c.updateCRC(byte(p))

	return c
}

func (e *emptySerializer) GetOutputMod() uint {
	return 1000
}

func (e *emptySerializer) SetXVal(uint, uint, *crc32Roller) {}

func (e *emptySerializer) SetPoly(poly uint) {
	e.poly = poly
}

func (e *emptySerializer) WriteOut() ([]byte, error) {
	return nil, nil
}

func (e *emptySerializer) ReadIn([]byte) error {
	return nil
}

type memSerializer struct {
	memoized [][256]uint
	mod      uint
	window   uint
	last     uint
	poly     uint
}

func (m *memSerializer) GetSerializedPos(window uint) uint {
	found := false
	for i := range m.memoized {
		if uint(i)*m.mod < window {
			found = true
			m.last = uint(i)
		}
	}

	if found {
		return m.last * m.mod
	}

	return 0
}

func (m *memSerializer) SetWindow(w uint) {
	m.window = w
	newMemoized := make([][256]uint, (w/m.mod)+1)

	if m.memoized != nil {
		for i, a := range m.memoized {
			for j := 0; j < 256; j++ {
				newMemoized[i][j] = a[j]
			}
		}
	}
	m.memoized = newMemoized
}

func (m *memSerializer) GetXVal(p int) *crc32Roller {
	if m.last == 0 {
		c := newCRC(m.poly, 0, nil, 1)
		c.updateCRC(byte(p))

		return c
	}

	crc := newCRC(m.poly, 0, nil, 1)
	crc.setCRC(m.memoized[m.last][p])

	return crc
}

func (m *memSerializer) GetOutputMod() uint {
	return m.mod
}

func (m *memSerializer) SetXVal(i uint, p uint, crc32 *crc32Roller) {
	i /= m.mod
	m.memoized[i][p] = crc32.crc32
}

func (m *memSerializer) SetPoly(poly uint) {
	m.poly = poly
}

func (m *memSerializer) WriteOut() ([]byte, error) {
	return json.Marshal(m.memoized)
}

func (m *memSerializer) ReadIn(data []byte) error {
	var memoized [][256]uint
	err := json.Unmarshal(data, &memoized)
	if err != nil {
		return err
	}

	m.memoized = memoized

	return nil
}

type jsonOutput struct {
	Mod      uint
	Memoized [][256]uint
}

type jsonSerializer struct {
	memSerializer
}

func (f *jsonSerializer) WriteOut() ([]byte, error) {
	out := jsonOutput{
		Mod:      f.mod,
		Memoized: f.memoized,
	}

	return json.Marshal(out)
}

func (f *jsonSerializer) ReadIn(data []byte) error {
	var in jsonOutput
	err := json.Unmarshal(data, &in)
	if err != nil {
		return err
	}

	f.mod = in.Mod
	f.memoized = in.Memoized

	return nil
}
