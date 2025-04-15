package main

import (
	"classify/cameras"
	"classify/jitter"
	"classify/video"
	"fmt"
	"net"
	"os"
	"strings"

	"gocv.io/x/gocv"
	// "github.com/pion/interceptor/pkg/jitterbuffer"
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
		server()
	}
}

// Comment
func stream() {
	cam1, err := video.NewVideo("./videos/vid.mp4")

	if err != nil {
		panic(err)
	}

	cameras.NewCameraFaker(cam1).Start(HOST, PORT)

	fmt.Println(fmt.Scanln())
}

type Payload struct {
	host  string
	frame *jitter.Frame
}

var payloads = make(chan *Payload)

type Stream struct {
	host   string
	buffer *jitter.Buffer
}

var streams = map[string]*Stream{}

// Comment
func server() {
	addr := net.UDPAddr{
		Port: PORT,
		IP:   net.ParseIP(HOST),
	}

	conn, err := net.ListenUDP("udp", &addr)

	if err != nil {
		panic(err)
	}

	window := gocv.NewWindow("UDP PACKETS")

	go run(conn)

	for payload := range payloads {
		if _, ok := streams[payload.host]; ok {
			mat, err := BytesToMat(payload.frame.Data)

			if err != nil {
				continue
			}

			window.IMShow(mat)
			window.WaitKey(1)
		}
	}

	conn.Close()
}

// Comment
func run(conn *net.UDPConn) {
	for {
		buffer := make([]byte, 2048)
		n, remoteAddr, err := conn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error reading packet:", err)
			continue
		}

		if n < 14 {
			fmt.Printf("Packet from %v too short: %d bytes\n", remoteAddr, n)
			continue
		}

		stream, ok := streams[remoteAddr.String()]

		if !ok {
			stream = &Stream{
				host:   remoteAddr.String(),
				buffer: jitter.NewBuffer(),
			}

			go Reader(stream)

			// Need struct for mutex
			streams[remoteAddr.String()] = stream
		}

		stream.buffer.Receive(buffer[:n])
	}
}

// Comment
func Reader(stream *Stream) {
	for {
		frame := stream.buffer.Read()

		if frame == nil {
			continue
		}

		payloads <- &Payload{host: stream.host, frame: frame}
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
