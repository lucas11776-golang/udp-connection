package server

import (
	"fmt"
	"net"
)

type PayloadCallback func(host string, payload []byte)

type Server struct {
	udp      *net.UDPConn
	messages []PayloadCallback
}

// Surveillance

// Comment
func Serve(host string, port int) *Server {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		panic(err)
	}

	return &Server{udp: conn}
}

// Comment
func (ctx *Server) Payload(callback PayloadCallback) {
	ctx.messages = append(ctx.messages, callback)
}

// Comment
func (ctx *Server) Listen() {
	for {
		buff := make([]byte, 1024*2)

		// Read incoming UDP packets.
		n, remoteAddr, err := ctx.udp.ReadFromUDP(buff)

		if err != nil {
			continue
		}

		for _, callback := range ctx.messages {
			callback(remoteAddr.String(), buff[:n])
		}
	}
}

// Comment
func (ctx *Server) Close() error {
	return ctx.udp.Close()
}
