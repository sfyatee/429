package main

import (
	"net"
	"bufio"
	"fmt"
	"os"
	"strings"
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

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("chat> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
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

		case "exit":
			fmt.Println("Exiting program.")
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}
