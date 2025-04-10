package video

import (
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
	"sync"
	"time"
)

const (
	END int64 = 1000000000000000000
)

type Frame struct {
	Total int
	Data  []byte
}

type BufferCallback func(frame *Frame)

// Dirty version of constructing packets
type Stream struct {
	Current int64
	Last    int64
	mutex   sync.Mutex
	// Will crash on long video overload :()
	Payloads map[int64]*Payload
	Buffers  []BufferCallback
}

type Payload struct {
	FramePerSecond int
	Fragments      int
	Packets        []*Packet
}

type Packet struct {
	Order int
	Data  []byte
}

// Comment
func NewStream() *Stream {
	stream := &Stream{
		Payloads: make(map[int64]*Payload),
	}

	// Start bufferring stream packets
	go stream.start()

	return stream
}

// Comment
func (ctx *Stream) Buffer(callback BufferCallback) {
	ctx.Buffers = append(ctx.Buffers, callback)
}

func Map(p []*Packet) []byte {
	b := []byte{}

	for _, p := range p {
		b = append(b, p.Data...)
	}

	return b
}

func (ctx *Stream) deletePayload(key int64) {
	ctx.mutex.Lock()
	delete(ctx.Payloads, key)
	ctx.mutex.Unlock()
}

// Comment
func (ctx *Stream) start() {
	tick := time.Tick(10 * time.Millisecond)

	for range tick {
		// Will skip frame :(
		if ctx.Current <= END {
			ctx.Current = ctx.Last
		}

		payload, ok := ctx.Payloads[ctx.Current]

		if !ok {
			ctx.deletePayload(ctx.Current)
			ctx.Current++
			continue
		}

		if len(payload.Packets) != payload.Fragments {
			ctx.deletePayload(ctx.Current)
			ctx.Current++
			continue
		}

		sort.Slice(payload.Packets, func(i, j int) bool {
			return payload.Packets[i].Order < payload.Packets[j].Order
		})

		py := slices.Clone(Map(payload.Packets))

		for _, callback := range ctx.Buffers {

			go callback(&Frame{
				Total: payload.FramePerSecond,
				Data:  py,
			})
		}

		ctx.deletePayload(ctx.Current)

		fmt.Println("PACKET:", ctx.Current, " SIZE:", len(py))

		ctx.Current++
	}
}

// Comment
func (ctx *Stream) Packet(p []byte) {
	if len(p) < 14 {
		return
	}

	packet := binary.BigEndian.Uint64(p[:8])
	size := binary.BigEndian.Uint16(p[8:10])
	order := binary.BigEndian.Uint16(p[10:12])
	fps := binary.BigEndian.Uint16(p[12:14])

	ctx.mutex.Lock()
	payload, ok := ctx.Payloads[int64(packet)]
	ctx.mutex.Unlock()

	if !ok {
		ctx.mutex.Lock()

		ctx.Payloads[int64(packet)] = &Payload{
			Fragments:      int(size),
			FramePerSecond: int(fps),
			Packets: []*Packet{
				{
					Order: int(order),
					Data:  p[14:],
				},
			},
		}

		ctx.Last = int64(packet)

		ctx.mutex.Unlock()

		return
	}

	payload.Packets = append(payload.Packets, &Packet{
		Order: int(order),
		Data:  p[14:],
	})

	// if payload.Fragments == len()
}
