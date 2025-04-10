package video

import (
	"errors"
	"fmt"

	"gocv.io/x/gocv"
)

type Video struct {
	img  gocv.Mat
	vCap *gocv.VideoCapture
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

type VideoAbstract interface {
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

func (ctx *Video) Close() {
	if err := ctx.img.Close(); err != nil {
		fmt.Printf("error: %v", err)
	}

	if err := ctx.vCap.Close(); err != nil {
		fmt.Printf("error: %v", err)
	}
}
