package mp4

import (
	"fmt"
	"os"
	"time"

	"gocv.io/x/gocv"
)

type MP4 struct {
	storage string
}

// Comment
func New(storage string) *MP4 {
	return &MP4{
		storage: storage,
	}
}

func BytesToMat(data []byte) (gocv.Mat, error) {
	mat, err := gocv.IMDecode(data, gocv.IMReadColor)

	if err != nil {
		return gocv.NewMat(), err
	}

	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("decoded mat is empty")
	}

	return mat, nil
}

// Coment
func (ctx *MP4) Record(t time.Time, frame []byte) error {
	_, err := BytesToMat(frame) // Valid is image....

	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%d.jpg", ctx.storage, t.UnixMilli())

	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	_, err = f.Write(frame)

	if err != nil {
		return err
	}

	return nil
}

// Comment
// TODO: store it in temp file (string, error).
func (ctx *MP4) Build(start time.Time, end time.Time, fps int, quality byte, width int, height int) ([]byte, error) {
	return nil, nil
}
