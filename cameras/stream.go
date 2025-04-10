package cameras

import (
	"encoding/binary"
	"fmt"
	"image"
	"net"

	"gocv.io/x/gocv"
)

type LiveStream interface {
	Reader(callback func(frame gocv.Mat, err error))
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

// Comment
func (ctx *LiveStreamFaker) packet(fragments uint16, packet uint64, order uint16, data []byte) []byte {
	size := make([]byte, 2)

	binary.BigEndian.PutUint16(size, fragments)

	pk := append([]byte{}, size...)

	no := make([]byte, 8)

	binary.BigEndian.PutUint64(no, packet)

	pk = append(pk, no...)

	ord := make([]byte, 2)

	binary.BigEndian.PutUint16(ord, order)

	pk = append(pk, ord...)

	return append(pk, data...)
}

const (
	MTU = 1400 - 12 // 1388
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

	index := 0

	video.Reader(func(frame gocv.Mat, err error) {
		if err != nil {
			return
		}

		err = gocv.Resize(frame, &frame, image.Pt(720, 360), 0, 0, gocv.InterpolationLinear)

		if err != nil {
			return
		}

		// Encode as JPEG to reduce size
		buff, err := gocv.IMEncode(".jpg", frame)

		if err != nil {
			return
		}

		payload := buff.GetBytes()

		fmt.Println("PACKET:", index, " SIZE:", len(payload))

		size := len(payload)
		div := size / MTU
		rim := size % MTU

		if rim > 0 {
			div += 1
		}

		for i := range div {
			s := i * MTU
			e := i*MTU + MTU

			if e < len(payload) {
				conn.Write(ctx.packet(uint16(div), uint64(index), uint16(i), payload[s:e]))
			} else {
				conn.Write(ctx.packet(uint16(div), uint64(index), uint16(i), payload[s:]))
			}
		}

		index++
	})
}
