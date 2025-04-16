package main

import (
	"classify/cameras"
	"classify/rtc"
	"classify/video"
	"context"
	"fmt"
	h "net/http"
	"os"
	"strings"

	"github.com/lucas11776-golang/http"
)

const (
	HOST = "127.0.0.1"
	PORT = 8080
)

func main() {
	if len(os.Args) < 2 {
		return
	}

	switch strings.ToLower(os.Args[1]) {
	case "stream":
		stream()
	case "server":
		go server()
		rtc.Server(HOST, PORT)
	}
}

// Comment
func stream() {
	cam1, err := video.NewVideo("./videos/vid.mp4")
	// cam1, err := video.NewCam(0)

	if err != nil {
		panic(err)
	}

	cameras.NewCameraFaker(cam1).Start(HOST, PORT)

	fmt.Println(fmt.Scanln())
}

func server() {
	server := http.Server(HOST, PORT).SetView("views", "html")

	server.Route().Group("videos", func(route *http.Router) {
		route.Group("{video}", func(route *http.Router) {
			route.Get("/", func(req *http.Request, res *http.Response) *http.Response {
				return res.View("video", http.ViewData{"video": "1"})
			})
		})
	})

	server.Route().Group("streams", func(route *http.Router) {
		route.Group("{id}", func(route *http.Router) {
			route.Get("/", func(req *http.Request, res *http.Response) *http.Response {
				stream := rtc.Streams.Filter(func(i interface{}) bool {
					return i.(*rtc.Payload).Host != ""
				}).Map(func(ctx context.Context, i interface{}) (interface{}, error) {
					return i.(*rtc.Payload).Frame.Data, nil
				}).Observe()

				req.Response.Writer.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

				end := req.Request.Context().Done()

				for frame := range stream {
					select {
					case <-end:
						return nil

					default:
						res.Writer.Write([]byte("--frame\r\n"))
						res.Writer.Write([]byte("Content-Type: image/jpeg\r\n"))

						res.Writer.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(frame.V.([]byte)))))

						res.Writer.Write(frame.V.([]byte))
						res.Writer.Write([]byte("\r\n"))

						if f, ok := res.Writer.(h.Flusher); ok {
							f.Flush()
						}
					}
				}

				return res
			})
		})
	})

	server.Listen()
}

// Can you show me a GO example of MJPEG frames for video stream?

// Want a quick example in Go or Python on how to serve MJPEG from those frames

// // Write multipart response
// w.Write([]byte("--frame\r\n"))
// w.Write([]byte("Content-Type: image/jpeg\r\n"))
// w.Write([]byte("Content-Length: " + string(len(buf)) + "\r\n\r\n"))
// w.Write(buf)
// w.Write([]byte("\r\n"))

// // Stream at ~30fps
// time.Sleep(33 * time.Millisecond)

// // Flush to client
// if f, ok := w.(http.Flusher); ok {
// 	f.Flush()
// }

// writer, err := gocv.VideoWriterFile("output.mp4", "avc1", 30, 640, 480, true)
// if err != nil {
// 	log.Fatal("Error opening video writer:", err)
// }
// defer writer.Close()
