package client

import (
	"net"
)

func Start(ip string, port string)(net.Conn, error) {
	address := ip + ":" + port

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
