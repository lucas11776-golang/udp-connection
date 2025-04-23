package file

/*
#include "./file.cpp"
*/
import "C"
import "unsafe"

// Comment
func SpliceFile(filename string, start int64, end int64) bool {
	str := C.CString(filename)

	spliced := C.splice(str, 0, 5)

	C.free(unsafe.Pointer(str))

	if spliced == 0 {
		return false
	}

	return true
}
