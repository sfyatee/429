package srv

import (
	"log"
	"fmt"
	"net"
)

func Start(port string) {
	listener, err := net.Listen("tcp", ":" + port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}		
	fmt.Println("Server started on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New connection from", conn.RemoteAddr())
		conn.Close()
	}
}
