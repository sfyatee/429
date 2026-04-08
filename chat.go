package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"chat/client"
	"chat/srv"

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
type Peer struct {
	ID int
	IP string
	Port string
	Conn net.Conn
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

	var peers []Peer
	nextID := 1

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
			fmt.Println("connect <IP> <port> - connect to another peer")
			fmt.Println("list - display all active connections")
			fmt.Println("terminate <id> - terminate a specific connection")
			fmt.Println("send <id> <message> - send a message to a specific connection")
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

			conn, err := client.Start(ip, destPort)
			if err != nil {
				fmt.Println("Error connecting to peer:", err)
				continue
			}
			peer := Peer{
				ID: nextID,
				IP: ip,
				Port: destPort,
				Conn: conn,
			}
			peers = append(peers, peer)
			fmt.Println("Connected to", ip, "on port", destPort)
			nextID++

		case "list":
			if len(peers) == 0 {
				fmt.Println("No active connections.")
				continue
			}
			fmt.Println("ID: IP address port number")
			for _, peer := range peers {
				fmt.Printf("%d: %s %s\n", peer.ID, peer.IP, peer.Port)
			}	

		case "terminate":
			if len(parts) != 2 {
				fmt.Println("Usage: terminate <connection id>")
				continue
			}

			idStr := parts[1]
			found := false

			for i, peer := range peers {
				if fmt. Sprintf("%d", peer.ID) == idStr {
					peer.Conn.Close()

					peers = append(peers[:i], peers[i+1:]...)
					fmt.Println("Connection has been terminated:", idStr)
					found = true
					break
				}
			}
			if !found {
				fmt.Println("Connection not found:", idStr)
			}

		case "send":
			if len(parts) < 3 {
				fmt.Println("Usage: send <connection id> <message>")
				continue
			}

			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid connection id")
				continue
			}

			msgText := strings.Join(parts[2:], " ")

			found := false
			for _, peer := range peers {
				if peer.ID == id {
					_, err := peer.Conn.Write([]byte(msgText + "\n"))
					if err != nil {
						fmt.Println("Error sending message")
					} else {
						fmt.Println("Message sent to connection", id)
					}
					found = true
					break
				}
			}

			if !found {
				fmt.Println("Connection not found:", id)
			}
		case "exit":
			for _, peer := range peers {
				peer.Conn.Close()
			}
			
			fmt.Println("Exiting program.")
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}
