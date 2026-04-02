package server

import (
	"fmt"
	"net"
	"strconv"
)

type Server struct {
	port int
	quit chan struct{}
}

func New(_port int) *Server {
	return &Server{port: _port}
}

func (s *Server) CreateServer() error {
	servicePort := ":" + strconv.Itoa(s.port)
	conn, err := net.Listen("tcp", servicePort)

	if err != nil {
		return err
	}

	defer conn.Close()

	for {
		newConn, err := conn.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(newConn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		select {
		case <-s.quit:
			return
		default:
			buff := make([]byte, 1024)

			n, err := conn.Read(buff)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(buff[:n])
			// parser
			// executor
			// response
			// write response
		}
	}
}
