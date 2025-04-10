package main

import (
	"classify/cameras"
	"classify/server"
	"classify/video"
	"fmt"
	"os"
	"strings"

	"gocv.io/x/gocv"
)

const (
	HOST = "127.0.0.1"
	PORT = 8080
)

func main() {
	if len(os.Args) >= 2 {
		switch strings.ToLower(os.Args[1]) {

		case "server":
			serverCall()

		case "stream":
			camerasCall()

		}
	}
}

// Comment
func camerasCall() {
	cam1, err := video.NewVideo("./videos/vid3.mp4")

	if err != nil {
		panic(err)
	}

	cameras.NewCameraFaker(cam1).Start(HOST, PORT)

	fmt.Println(fmt.Scanln())
}

func (ctx *LiveStream) BytesToMat(data []byte) (gocv.Mat, error) {
	mat, err := gocv.IMDecode(data, gocv.IMReadColor)

	if err != nil {
		return gocv.NewMat(), err
	}

	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("decoded mat is empty")
	}

	return mat, nil
}

type LiveStream struct {
	host   string
	stream *video.Stream
}

// Comment
func NewStream(ip string) *LiveStream {
	c := &LiveStream{
		host:   ip,
		stream: video.NewStream(),
	}

	go c.buffer()

	return c
}

type FrameView struct {
	Host string
	Mat  gocv.Mat
}

var FrameChan = make(chan *FrameView)

// Comment
func (ctx *LiveStream) buffer() {
	ctx.stream.Buffer(func(data []byte) {
		mat, err := ctx.BytesToMat(data)

		if err != nil {
			return
		}

		FrameChan <- &FrameView{
			Host: ctx.host,
			Mat:  mat,
		}
	})
}

// Comment
func (ctx *LiveStream) Show(window *gocv.Window, data []byte) {
	mat, err := ctx.BytesToMat(data)

	if err != nil {
		return
	}

	window.IMShow(mat)
}

type Window struct {
	window *gocv.Window
}

// Comment
func (ctx *Window) Show(mat gocv.Mat) {
	ctx.window.IMShow(mat)
	ctx.window.WaitKey(1)
}

// Comment
func serverCall() {
	window := &Window{window: gocv.NewWindow("UDP Video Stream")}

	server := server.Serve(HOST, PORT)

	cams := map[string]*LiveStream{}

	server.Payload(func(host string, packet []byte) {
		cam, ok := cams[host]

		if !ok {
			cams[host] = NewStream(host)

			cam = cams[host]
		}

		cam.stream.Packet(packet)
	})

	go server.Listen()

	for f := range FrameChan {
		window.Show(f.Mat)
	}

	window.window.Close()
}
