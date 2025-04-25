package rtc

import (
	"classify/ffmpeg/mp4"
	"classify/jitter"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/reactivex/rxgo/v2"
	"gocv.io/x/gocv"
)

const STORAGE = "./storage"

type Payload struct {
	Host  string
	Frame *jitter.Frame
}

var payloads = make(chan *Payload)
var payloadsRx = make(chan rxgo.Item)

type Stream struct {
	storage string
	host    string
	buffer  *jitter.Buffer
}

// Comment
func NewStream(host string) (*Stream, error) {
	cwd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	return &Stream{
		storage: fmt.Sprintf("%s/storage", cwd),
		host:    host,
		buffer:  jitter.NewBuffer(),
	}, nil
}

// Comment
func (ctx *Stream) Record(frame []byte) error {

	// err := avi.Open(fmt.Sprintf("%s/%s", ctx.storage, ctx.filename("mp4")), 60, 1, 720, 360, ctx.storage).Add(frame)

	// err := v.Add(frame)

	// if err != nil {
	// 	return err
	// }

	return nil
}

func (ctx *Stream) filename(ext string) string {
	t := time.Now()
	return fmt.Sprintf("%d-%s-%d.%s", t.Day(), t.Month().String(), t.Year(), ext)
}

var streams = map[string]*Stream{}

var Streams = rxgo.FromChannel(payloadsRx)

// Comment
func Server(host string, port int) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	conn, err := net.ListenUDP("udp", &addr)

	if err != nil {
		panic(err)
	}

	go run(conn)

	for payload := range payloads {
		if _, ok := streams[payload.Host]; ok {
			go func() { payloadsRx <- rxgo.Of(payload) }()
			go mp4.New(STORAGE).Record(time.UnixMilli(payload.Frame.Timestamp), payload.Frame.Data)
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
			stream, err = NewStream(remoteAddr.String())

			if err != nil {
				fmt.Println("Error", err)
				continue
			}

			go Reader(stream)

			// Need struct for mutex for conc, but TEMP
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

		payloads <- &Payload{
			Host:  stream.host,
			Frame: frame,
		}
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
