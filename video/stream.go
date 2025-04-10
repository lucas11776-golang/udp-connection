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

// Dirty version of constructing packets
type Stream struct {
	Current int64
	Last    int64
	mutex   sync.Mutex
	// Will crash on long video overload :()
	Payloads map[int64]*Payload
	Buffers  []func(data []byte)
}

type Payload struct {
	Fragments int
	Packets   []*Packet
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
func (ctx *Stream) Buffer(callback func(data []byte)) {
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
	delete(ctx.Payloads, ctx.Current)
	ctx.mutex.Unlock()
}

// Comment
func (ctx *Stream) start() {
	tick := time.Tick(10 * time.Millisecond)

	for range tick {
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
			go callback(py)
		}

		ctx.deletePayload(ctx.Current)

		fmt.Println("PACKET:", ctx.Current, " SIZE:", len(py))

		ctx.Current++
	}
}

// Comment
func (ctx *Stream) Packet(p []byte) {
	if len(p) < 12 {
		return
	}

	size := binary.BigEndian.Uint16(p[:2])
	packet := binary.BigEndian.Uint64(p[2:10])
	order := binary.BigEndian.Uint16(p[10:12])

	ctx.mutex.Lock()
	payload, ok := ctx.Payloads[int64(packet)]
	ctx.mutex.Unlock()

	if !ok {
		ctx.mutex.Lock()

		ctx.Payloads[int64(packet)] = &Payload{
			Fragments: int(size),
			Packets: []*Packet{
				{
					Order: int(order),
					Data:  p[12:],
				},
			},
		}

		ctx.Last = int64(packet)

		ctx.mutex.Unlock()

		return
	}

	payload.Packets = append(payload.Packets, &Packet{
		Order: int(order),
		Data:  p[12:],
	})

	// if payload.Fragments == len()
}
