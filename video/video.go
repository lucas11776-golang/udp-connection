package video

import (
	"errors"
	"math"

	"gocv.io/x/gocv"
)

type Video struct {
	img  gocv.Mat
	vCap *gocv.VideoCapture
	fps  int
}

// Comment
func NewVideo(path string) (*Video, error) {
	vCap, err := gocv.VideoCaptureFile(path)
	// vCap, err := gocv.VideoCaptureDevice(0)

	if err != nil {
		return nil, err
	}

	mat := gocv.NewMat()

	return &Video{
		img:  mat,
		vCap: vCap,
		fps:  int(math.Ceil(vCap.Get(gocv.VideoCaptureFPS))),
	}, nil
}

// Comment
func NewWebCam(device int) (*Video, error) {
	vCap, err := gocv.VideoCaptureDevice(device)

	if err != nil {
		return nil, err
	}

	mat := gocv.NewMat()

	return &Video{
		img:  mat,
		vCap: vCap,
	}, nil
}

type StreamAbstract interface {
	Reader(callback func(frame gocv.Mat, err error))
	Close()
}

func (ctx *Video) Reader(callback func(frame gocv.Mat, err error)) {
	for {
		if ok := ctx.vCap.Read(&ctx.img); !ok {
			callback(gocv.Mat{}, errors.New("failed to read video"))

			break
		}

		if ctx.img.Empty() {
			continue
		}

		callback(ctx.img, nil)
	}
}

// Comment
func (ctx *Video) Fps() int {
	return ctx.fps
}

// Commet
func (ctx *Video) Close() {
	if err := ctx.img.Close(); err != nil {
	}

	if err := ctx.vCap.Close(); err != nil {
	}
}
