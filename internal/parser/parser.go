package parser

import (
	"bufio"
	"fmt"
	"net"
)

type Parser struct {
	connection net.Conn
}

func New(conn net.Conn) *Parser {
	return &Parser{connection: conn}
}

func (p *Parser) Start() {
	reader := bufio.NewReaderSize(p.connection, 4096)

	for {
		requestLine, err := reader.ReadString('\n')

		if err != nil {
			panic(fmt.Sprintf("Error ", err))
		}
		fmt.Println(requestLine)
	}
}
