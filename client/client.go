package client

import (
	"fmt"
	"net"
)

func Start(ip string, port string) {
	address := ip + ":" + port

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}

	fmt.Println("Connected to", ip, "on port", port)

	conn.Close()
}
