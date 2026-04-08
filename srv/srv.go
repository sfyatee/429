package srv

import (
	// "log"
	"bufio"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed by", conn.RemoteAddr())
			conn.Close()
			return
		}
		fmt.Println("Message from", conn.RemoteAddr())
		fmt.Print("Message: ", message)
		fmt.Print("chat> ")
	}
}
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
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New connection from", conn.RemoteAddr())
		go handleConnection(conn)
	}
}
