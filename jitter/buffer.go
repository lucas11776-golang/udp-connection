package jitter

import (
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Fragment struct {
	position int
	data     []byte
}

// Comment
func (ctx *Fragment) Position() int {
	return ctx.position
}

// Comment
func (ctx *Fragment) Data() []byte {
	return ctx.data
}

type Packet struct {
	time      time.Time
	position  int64
	fps       int
	total     int
	fragments map[int]*Fragment
}

// Comment
func (ctx *Packet) Ready() bool {
	return len(ctx.fragments) == ctx.total
}

// Comment
func (ctx *Packet) Position() int64 {
	return ctx.position
}

// Comment
func (ctx *Packet) Total() int {
	return ctx.total
}

// Comment
func (ctx *Packet) Frames() int {
	return ctx.fps
}

// Comment
func (ctx *Packet) Data() []byte {
	var positions []int
	var data []byte

	for pos := range ctx.fragments {
		positions = append(positions, pos)
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i] < positions[j]
	})

	for _, pos := range positions {
		data = append(data, ctx.fragments[pos].Data()...)
	}

	return data
}

type Buffer struct {
	packets      map[int64]*Packet
	mutexPackets sync.Mutex
	stream       *Stream
	timeout      time.Duration
}

// Comment
func NewBuffer() *Buffer {
	buffer := &Buffer{
		packets: make(map[int64]*Packet),
		timeout: time.Millisecond * 500,
		stream:  &Stream{},
	}

	go buffer.cleanup()

	return buffer
}

func (ctx *Buffer) Receive(buffer []byte) {
	if len(buffer) < 14 {
		return
	}

	ctx.payload(
		int64(binary.BigEndian.Uint64(buffer[:8])),
		int(binary.BigEndian.Uint16(buffer[8:10])),
		int(binary.BigEndian.Uint16(buffer[10:12])),
		int(binary.BigEndian.Uint16(buffer[12:14])),
		buffer[14:],
	)
}

// Comment
func (ctx *Buffer) payload(packetNumber int64, fragmentNumber int, fragmentPosition int, fps int, data []byte) {
	ctx.mutexPackets.Lock()
	defer ctx.mutexPackets.Unlock()

	packet, ok := ctx.packets[packetNumber]

	if !ok {
		packet = &Packet{
			position:  packetNumber,
			total:     fragmentNumber,
			fps:       fps,
			fragments: make(map[int]*Fragment),
			time:      time.Now(),
		}

		ctx.packets[packetNumber] = packet
	}

	packet.fragments[fragmentPosition] = &Fragment{
		position: fragmentPosition,
		data:     data,
	}

	if len(packet.fragments) >= packet.total {
		ctx.stream.Store(packet)
		delete(ctx.packets, packetNumber)
	}
}

// Comment
func (ctx *Buffer) cleanup() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		ctx.mutexPackets.Lock()
		for packetNum, pkt := range ctx.packets {
			if now.Sub(pkt.time) > ctx.timeout {
				delete(ctx.packets, packetNum)
			}
		}
		ctx.mutexPackets.Unlock()
	}
}

type Stream struct {
	mutex   sync.Mutex
	packets []*Packet
}

// Comment
func (ctx *Stream) Get() *Packet {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if len(ctx.packets) == 0 {
		return nil
	}

	packet := ctx.packets[0]
	ctx.packets = ctx.packets[1:]

	return packet
}

func (ctx *Stream) Frame() *Frame {
	packet := ctx.Get()

	if packet == nil {
		return nil
	}

	var positions []int

	fmt.Println("PACKET", packet.position, "FRAME LEFT", len(ctx.packets))

	for pos := range packet.fragments {
		positions = append(positions, pos)
	}

	sort.Ints(positions)

	var data []byte

	for _, pos := range positions {
		data = append(data, packet.fragments[pos].data...)
	}

	return &Frame{Rate: packet.fps, Data: data}
}

// Comment
func (ctx *Stream) Store(packet *Packet) {
	// May receive old packet when reading :(
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.packets = append(ctx.packets, packet)

	sort.Slice(ctx.packets, func(i, j int) bool {
		return ctx.packets[i].position < ctx.packets[j].position
	})
}

type Frame struct {
	Rate int
	Data []byte
}

// comment
func (ctx *Buffer) Read() *Frame {
	return ctx.stream.Frame()
}
