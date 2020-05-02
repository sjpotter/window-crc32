package window_crc32

import (
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRC32(t *testing.T) {
	var buffer []byte
	for i := 0; i < 300; i++ {
		j := 11 + i*31 + i/17
		buffer = append(buffer, byte(j))
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer)

	myCrc32 := NewCRC(CrcPoly, 0, nil)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32WindowWithoutRoll(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[0:100])

	myCrc32 := NewCRC(CrcPoly, 100, nil)
	for _, b := range buffer[:100] {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32Rolling(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[200:300])

	myCrc32 := NewCRC(CrcPoly, 100, nil)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32RollingEmptySerializer(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[200:300])

	myCrc32 := NewCRC(CrcPoly, 100, &emptySerializer{})
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32RollingMemSerializer(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[200:300])

	m := &memSerializer{mod: 75}

	myCrc32 := NewCRC(CrcPoly, 100, m)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())

	myCrc32 = NewCRC(CrcPoly, 100, m)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32RollingMemSerializerWriteRead(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[200:300])

	m := &memSerializer{mod: 75}

	myCrc32 := NewCRC(CrcPoly, 100, m)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())

	data, err := m.WriteOut()
	assert.Nil(t, err)

	m = &memSerializer{mod: 75}
	err = m.ReadIn(data)
	assert.Nil(t, err)

	goCrc32 = crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[150:300])

	myCrc32 = NewCRC(CrcPoly, 150, m)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32RollingFileSerializer(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[200:300])

	m := NewJsonSerializer(75)

	myCrc32 := NewCRC(CrcPoly, 100, m)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())

	data, err := m.WriteOut()
	assert.Nil(t, err)

	m = NewJsonSerializer(0)
	err = m.ReadIn(data)
	assert.Nil(t, err)

	goCrc32 = crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[150:300])

	myCrc32 = NewCRC(CrcPoly, 150, m)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}

func TestCRC32RollingMemSerializerThreaded(t *testing.T) {
	buffer := make([]byte, 300)

	for i := 0; i < 300; i++ {
		buffer[i] = byte(11 + i*31 + i/17)
	}

	goCrc32 := crc32.NewIEEE()
	_, _ = goCrc32.Write(buffer[200:300])

	m := &memSerializer{mod: 75}

	myCrc32 := NewCRCThreaded(CrcPoly, 100, m, 8)
	for _, b := range buffer {
		myCrc32.UpdateCRC32(b)
	}

	assert.Equal(t, uint(goCrc32.Sum32()), myCrc32.Finish())
}
