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

func video(vc *gocv.VideoCapture) *Video {
	mat := gocv.NewMat()

	return &Video{
		img:  mat,
		vCap: vc,
		fps:  int(math.Ceil(vc.Get(gocv.VideoCaptureFPS))),
	}
}

// Comment
func NewCam(device int) (*Video, error) {
	vc, err := gocv.VideoCaptureDevice(device)

	if err != nil {
		return nil, err
	}

	return video(vc), nil
}

// Comment
func NewVideo(path string) (*Video, error) {
	vc, err := gocv.VideoCaptureFile(path)

	if err != nil {
		return nil, err
	}

	return video(vc), nil
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
