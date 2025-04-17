package cameras

import (
	"encoding/binary"
	"fmt"
	"image"
	"net"
	"time"

	"gocv.io/x/gocv"
)

type LiveStream interface {
	Reader(callback func(frame gocv.Mat, err error))
	Fps() int
	Close()
}

type LiveStreamFaker struct {
	videos []LiveStream
}

// Comment
func NewCameraFaker(videos ...LiveStream) *LiveStreamFaker {
	return &LiveStreamFaker{videos: videos}
}

// Comment
func (ctx *LiveStreamFaker) Start(host string, port int) {
	for _, video := range ctx.videos {
		ctx.fakeVideoFeed(host, port, video)
	}
}

// Number -> 8
// Size   -> 2
// Order  -> 2
// FPS    -> 2
// Data   -> [14:]

// Comment
func (ctx *LiveStreamFaker) packet(frame int, fragments uint16, packet uint64, order uint16, data []byte) []byte {
	pk := []byte{}

	// Packet No.
	no := make([]byte, 8)

	binary.BigEndian.PutUint64(no, packet)

	pk = append(pk, no...)

	// Packet Chuck Size
	size := make([]byte, 2)

	binary.BigEndian.PutUint16(size, fragments)

	pk = append(pk, size...)

	// Packet Order
	ord := make([]byte, 2)

	binary.BigEndian.PutUint16(ord, order)

	pk = append(pk, ord...)

	// Packet FPS
	fr := make([]byte, 2)

	binary.BigEndian.PutUint16(fr, uint16(frame))

	pk = append(pk, fr...)

	return append(pk, data...)
}

const (
	MTU = 1400 - 14 // 1388
)

// Comment
func (ctx *LiveStreamFaker) fakeVideoFeed(host string, port int, video LiveStream) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)

	if err != nil {
		panic(err)
	}

	video.Reader(func(frame gocv.Mat, err error) {

		// if index > 0 {
		// 	os.Exit(0)
		// 	return
		// }

		if err != nil {
			return
		}

		err = gocv.Resize(frame, &frame, image.Pt(720, 360), 0, 0, gocv.InterpolationLinear)

		if err != nil {
			return
		}

		buff, err := gocv.IMEncode(".jpg", frame)

		if err != nil {
			return
		}

		payload := buff.GetBytes()

		size := len(payload)
		div := size / MTU
		rim := size % MTU

		if rim > 0 {
			div += 1
		}

		timestamp := time.Now().UnixMilli()

		for i := range div {
			start := i * MTU
			end := i*MTU + MTU

			if end < len(payload) {
				conn.Write(ctx.packet(video.Fps(), uint16(div), uint64(timestamp), uint16(i), payload[start:end]))
			} else {
				conn.Write(ctx.packet(video.Fps(), uint16(div), uint64(timestamp), uint16(i), payload[start:]))
			}

			fmt.Println("PACKET", timestamp, i)
		}
	})
}
