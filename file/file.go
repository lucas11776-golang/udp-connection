package file

/*
#include "./file.cpp"
*/
import "C"
import (
	"os"
	"unsafe"
)

// Comment
func Exists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}

// Comment
func SpliceFile(filename string, start int64, end int64) bool {
	str := C.CString(filename)

	spliced := C.splice(str, C.double(start), C.double(end))

	C.free(unsafe.Pointer(str))

	if spliced == 0 {
		return false
	}

	return true
}

// Comment
func SplicePrefix(filename string, prefix string, offset int64) bool {
	str := C.CString(filename)
	pre := C.CString(prefix)

	spliced := C.splicePrefix(str, pre, C.double(offset))

	C.free(unsafe.Pointer(str))
	C.free(unsafe.Pointer(pre))

	if spliced == 0 {
		return false
	}

	return true
}

// Comment
func Join(dst string, src string) bool {
	d := C.CString(dst)
	s := C.CString(src)

	joined := C.join(d, s)

	C.free(unsafe.Pointer(d))
	C.free(unsafe.Pointer(s))

	if joined == 0 {
		return false
	}

	return true
}

// ffmpeg.exe -f image2 -framerate 25 -pattern_type sequence -start_number 1234 -r 3 -i Imgp%04d.jpg -s 720x480 test.avi

// ffmpeg -f image2 -framerate 30 -pattern_type glob -i './pictures/*.jpg' -r 3 -s 720x360 test.avi

// ffmpeg -f image2 -framerate 30 -pattern_type glob -i './pictures/*.jpg' \-c:v libx264 -pix_fmt yuv420p out.avi

// 720, 360

// ffmpeg -framerate 30 -i ./pictures/*.jpg -c:v mjpeg -q:v 2 output.avi
// ffmpeg -framerate 30 -i ./pictures/%d.jpg -c:v mjpeg -q:v 2 output.avi
