package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"chat/srv"
	"chat/client"
)

func usage() {
	fmt.Print("")
	os.Exit(2)
}

func getMyIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
	// help
	// myip
	// myport
	// connect
	// list
	// terminate
	// send
	// exit
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run chat.go <port>")
		return
	}
	port := os.Args[1]

	go srv.Start(port) // starts server so server can listen while command prompt is active

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("chat> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("help - display this help message")
			fmt.Println("myip - display the IP address of this machine")
			fmt.Println("myport - display the port number this program is listening on")
			fmt.Println("exit - exit the program")

		case "myip":
			fmt.Println(getMyIP())

		case "myport":
			fmt.Println(port)

		case "connect":
			if len(parts) != 3 {
				fmt.Println("Usage: connect <IP> <port>")
				continue
			}
			ip := parts[1]
			destPort := parts[2]
			client.Start(ip, destPort)

		case "exit":
			fmt.Println("Exiting program.")
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}
