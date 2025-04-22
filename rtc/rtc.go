package rtc

import (
	"classify/jitter"
	"classify/mjpeg"
	"fmt"
	"net"
	"time"

	"github.com/reactivex/rxgo/v2"
	"gocv.io/x/gocv"
)

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
	writer  *gocv.VideoWriter
	// avi     mjpeg.AviWriter
}

// Comment
func NewStream(host string) (*Stream, error) {
	stream := &Stream{
		storage: "./storage",
		host:    host,
		buffer:  jitter.NewBuffer(),
	}

	// writer, err := gocv.VideoWriterFile("./"+stream.storage+"/avi-"+stream.filename("mp4"), "mp4v", 60, 720, 360, true)

	// if err != nil {
	// 	fmt.Println("Error --> :", err)
	// 	return nil, err
	// }

	// stream.writer = writer

	// avi, err := mjpeg.NewOrExisting("./"+stream.storage+"/avi-"+stream.filename("avi"), 720, 360, 60)

	// if err != nil {
	// 	return nil, err
	// }

	// stream.avi = avi

	return stream, nil
}

// Comment
func (ctx *Stream) Record(frame [][]byte) {
	_, err := mjpeg.New(ctx.storage, frame, 720, 360, 60, ctx.filename("avi"))

	if err != nil {

		fmt.Println("ERRROR", err)

		return
	}
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

	window := gocv.NewWindow("UDP CONNECTION")

	index := 0

	frames := [][]byte{}

	for payload := range payloads {
		if stream, ok := streams[payload.Host]; ok {
			go func() { payloadsRx <- rxgo.Of(payload) }()

			mat, err := BytesToMat(payload.Frame.Data)

			if err != nil {
				continue
			}

			// TODO: testing - 5seconds video clip

			frames = append(frames, payload.Frame.Data)

			if index == (60 * 5) {
				stream.Record(frames)
				return
			}

			index++

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

		payloads <- &Payload{Host: stream.host, Frame: frame}
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
