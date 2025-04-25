package main

import (
	"classify/cameras"
	"classify/rtc"
	"classify/video"
	"context"
	"fmt"
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

						// if f, ok := res.Writer.(h.Flusher); ok {
						// 	f.Flush()
						// }
					}
				}

				return res
			})
		})
	})

	server.Listen()
}

/******************* TESTING *********/
