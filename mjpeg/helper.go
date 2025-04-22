package mjpeg

import (
	"encoding/binary"
	"os"
)

// Write data to file
func Write(f *os.File, data []byte) error {
	_, err := f.Write(data)
	return err
}

// WriteStr writes a string to the file.
func WriteStr(f *os.File, s string) error {
	_, err := f.WriteString(s)
	return err
}

// writeInt16 writes a 16-bit int value to the index file.
func writeInt16(f *os.File, n int16) error {
	buff := make([]byte, 2)
	binary.LittleEndian.PutUint16(buff, uint16(n))
	_, err := f.Write(buff)
	return err
}

// writeInt32 writes a 32-bit int value to the file.
func WriteInt32(f *os.File, n int32) error {
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(n))
	_, err := f.Write(buff)
	return err
}

// finalizeLengthField finalizes the last length field.
func FinalizeLengthField(f *os.File, lengthFields []int64) ([]int64, error) {
	pos, err := CurrentPos(f)

	if err != nil {
		return nil, err
	}

	numLenFs := len(lengthFields)

	if numLenFs == 0 {
		return nil, errImproperState
	}

	Seek(f, lengthFields[numLenFs-1], 0)
	lengthFields = lengthFields[:numLenFs-1]

	cPos, err := CurrentPos(f)

	if err != nil {
		return nil, err
	}

	WriteInt32(f, int32(pos-cPos-4))

	// Seek "back" but align to a 2-byte boundary
	if pos&0x01 != 0 {
		pos++
	}

	Seek(f, pos, 0)

	return lengthFields, err
}

// seek seeks the AVI file.
func Seek(f *os.File, offset int64, whence int) (int64, error) {
	return f.Seek(offset, whence)
}

// currentPos returns the current file position of the AVI file.
func CurrentPos(f *os.File) (int64, error) {
	return Seek(f, 0, 1) // Seek relative to current pos
}

// writeLengthField writes an empty int field to the avi file, and saves
// the current file position as it will be filled later.
func WriteLengthField(f *os.File, lengthFields []int64) ([]int64, error) {
	pos, err := CurrentPos(f)

	if err != nil {
		return nil, err
	}

	lengthFields = append(lengthFields, pos)

	err = WriteInt32(f, 0)

	if err != nil {
		return nil, err
	}

	return lengthFields, err
}
