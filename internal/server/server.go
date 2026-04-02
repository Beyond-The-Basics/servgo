package server

import (
	"net"
	"strconv"

	"github.com/Mohamed-Moumni/servgo/internal/parser"
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
			parser := parser.New(conn)
			parser.Start()
			// parser
			// executor
			// response
			// write response
		}
	}
}
