package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"chat/client"
	"chat/srv"
)

type Peer struct {
	ID   int
	IP   string
	Port string
	Link *srv.PeerConn
}

type ChatApp struct {
	listenPort string
	listener   net.Listener

	mu      sync.Mutex
	peers   map[int]*Peer
	nextID  int
	closing bool
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <port>")
		return
	}

	port := strings.TrimSpace(os.Args[1])
	if _, err := validatePort(port); err != nil {
		fmt.Println("Invalid listening port:", err)
		return
	}

	app := &ChatApp{
		listenPort: port,
		peers:      make(map[int]*Peer),
		nextID:     1,
	}

	if err := app.startServer(); err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	app.runShell()
}

func (a *ChatApp) startServer() error {
	listener, err := srv.Start(a.listenPort)
	if err != nil {
		return err
	}
	a.listener = listener

	go srv.AcceptLoop(listener, a.listenPort, srv.Handlers{
		IsClosing:        a.isClosing,
		ValidateIncoming: a.validateIncomingPeer,
		OnConnect:        a.onIncomingConnect,
		OnMessage:        a.onPeerMessage,
		OnDisconnect:     a.onPeerDisconnect,
		OnAcceptError: func(err error) {
			fmt.Println("\nError:", err)
			fmt.Print("chat> ")
		},
	})
	return nil
}

func (a *ChatApp) runShell() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("chat> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			a.exit()
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		a.handleCommand(line)
		if a.isClosing() {
			return
		}
	}
}

func (a *ChatApp) handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	switch strings.ToLower(parts[0]) {
	case "help":
		printHelp()
	case "myip":
		fmt.Println("The IP address is", getMyIP())
	case "myport":
		fmt.Println("The program runs on port number", a.listenPort)
	case "connect":
		if len(parts) != 3 {
			fmt.Println("Usage: connect <destination> <port>")
			return
		}
		if err := a.connect(parts[1], parts[2]); err != nil {
			fmt.Println(err)
		}
	case "list":
		a.list()
	case "terminate":
		if len(parts) != 2 {
			fmt.Println("Usage: terminate <connection id>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println("Invalid connection id")
			return
		}
		if err := a.terminate(id, true); err != nil {
			fmt.Println(err)
		}
	case "send":
		if len(parts) < 3 {
			fmt.Println("Usage: send <connection id> <message>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println("Invalid connection id")
			return
		}
		msg := strings.Join(parts[2:], " ")
		if err := a.send(id, msg); err != nil {
			fmt.Println(err)
		}
	case "exit":
		a.exit()
	default:
		fmt.Println("Unknown command. Type 'help' for available commands.")
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("help - display this help message")
	fmt.Println("myip - display the IP address of this machine")
	fmt.Println("myport - display the port number this program is listening on")
	fmt.Println("connect <IP> <port> - connect to another peer")
	fmt.Println("list - display all active connections")
	fmt.Println("terminate <id> - terminate a specific connection")
	fmt.Println("send <id> <message> - send a message to a specific connection")
	fmt.Println("exit - close all connections and exit the program")
}

func (a *ChatApp) connect(ip, port string) error {
	if err := validateIPv4(ip); err != nil {
		return err
	}
	if _, err := validatePort(port); err != nil {
		return err
	}
	if err := a.validateIncomingPeer(ip, port); err != nil {
		return err
	}

	session, err := client.Start(ip, port, a.listenPort)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	peer := &Peer{
		IP:   session.RemoteIP,
		Port: session.RemotePort,
		Link: &srv.PeerConn{Conn: session.Conn, Reader: session.Reader, RemoteIP: session.RemoteIP, RemotePort: session.RemotePort},
	}
	id := a.addPeer(peer)
	fmt.Printf("The connection to peer %s is successfully established.\n", ip)
	fmt.Printf("Connected as id %d\n", id)

	go a.readOutgoingPeer(peer)
	return nil
}

func (a *ChatApp) readOutgoingPeer(peer *Peer) {
	srv.MonitorPeer(peer.Link, srv.Handlers{
		OnMessage:    a.onPeerMessage,
		OnDisconnect: a.onPeerDisconnect,
	})
}

func (a *ChatApp) validateIncomingPeer(ip, port string) error {
	if _, err := validatePort(port); err != nil {
		return err
	}
	if a.isSelfConnection(ip, port) {
		return errors.New("self connection is not allowed")
	}
	if a.hasPeer(ip, port) {
		return errors.New("duplicate connection: already connected to that peer")
	}
	return nil
}

func (a *ChatApp) onIncomingConnect(link *srv.PeerConn) {
	peer := &Peer{IP: link.RemoteIP, Port: link.RemotePort, Link: link}
	id := a.addPeer(peer)
	fmt.Printf("\nThe connection to peer %s is successfully established.\n", link.RemoteIP)
	fmt.Printf("Connected as id %d\n", id)
	fmt.Print("chat> ")
}

func (a *ChatApp) onPeerMessage(link *srv.PeerConn, msg string) {
	if msg == "" {
		fmt.Printf("\nReceived unrecognized data from %s\n", link.RemoteIP)
		fmt.Print("chat> ")
		return
	}
	fmt.Printf("\nMessage received from %s\n", link.RemoteIP)
	fmt.Printf("Sender's Port: %s\n", link.RemotePort)
	fmt.Printf("Message: \"%s\"\n", msg)
	fmt.Print("chat> ")
}

func (a *ChatApp) onPeerDisconnect(link *srv.PeerConn, graceful bool) {
	removed := a.removePeerByConn(link.Conn)
	_ = link.Conn.Close()
	if removed == nil || a.isClosing() {
		return
	}
	if graceful {
		fmt.Printf("\nPeer %s terminated the connection.\n", removed.IP)
	} else {
		fmt.Printf("\nConnection closed by %s\n", removed.IP)
	}
	fmt.Print("chat> ")
}

func (a *ChatApp) list() {
	peers := a.snapshotPeers()
	if len(peers) == 0 {
		fmt.Println("No active connections.")
		return
	}
	fmt.Println("id: IP address Port No.")
	for _, peer := range peers {
		fmt.Printf("%d: %s %s\n", peer.ID, peer.IP, peer.Port)
	}
}

func (a *ChatApp) send(id int, msg string) error {
	if len(msg) > 100 {
		return errors.New("message must be 100 characters or fewer")
	}
	peer := a.getPeer(id)
	if peer == nil {
		return fmt.Errorf("connection not found: %d", id)
	}
	if err := client.SendMessage(peer.Link.Conn, msg); err != nil {
		a.removePeer(id)
		_ = peer.Link.Conn.Close()
		return fmt.Errorf("failed to send message to %d", id)
	}
	fmt.Printf("Message sent to %d\n", id)
	return nil
}

func (a *ChatApp) terminate(id int, notify bool) error {
	peer := a.removePeer(id)
	if peer == nil {
		return fmt.Errorf("connection not found: %d", id)
	}
	if notify {
		_ = client.SendTerminate(peer.Link.Conn)
	}
	_ = peer.Link.Conn.Close()
	fmt.Printf("Connection %d terminated.\n", id)
	return nil
}

func (a *ChatApp) exit() {
	a.mu.Lock()
	if a.closing {
		a.mu.Unlock()
		return
	}
	a.closing = true
	peers := make([]*Peer, 0, len(a.peers))
	for _, peer := range a.peers {
		peers = append(peers, peer)
	}
	a.peers = make(map[int]*Peer)
	a.mu.Unlock()

	for _, peer := range peers {
		_ = client.SendTerminate(peer.Link.Conn)
		_ = peer.Link.Conn.Close()
	}
	if a.listener != nil {
		_ = a.listener.Close()
	}
	fmt.Println("Exiting program.")
}

func (a *ChatApp) addPeer(peer *Peer) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	peer.ID = a.nextID
	a.peers[peer.ID] = peer
	a.nextID++
	return peer.ID
}

func (a *ChatApp) removePeer(id int) *Peer {
	a.mu.Lock()
	defer a.mu.Unlock()
	peer := a.peers[id]
	if peer != nil {
		delete(a.peers, id)
	}
	return peer
}

func (a *ChatApp) removePeerByConn(conn net.Conn) *Peer {
	a.mu.Lock()
	defer a.mu.Unlock()
	for id, peer := range a.peers {
		if peer.Link != nil && peer.Link.Conn == conn {
			delete(a.peers, id)
			return peer
		}
	}
	return nil
}

func (a *ChatApp) getPeer(id int) *Peer {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.peers[id]
}

func (a *ChatApp) hasPeer(ip, port string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, peer := range a.peers {
		if peer.IP == ip && peer.Port == port {
			return true
		}
	}
	return false
}

func (a *ChatApp) snapshotPeers() []*Peer {
	a.mu.Lock()
	defer a.mu.Unlock()
	peers := make([]*Peer, 0, len(a.peers))
	for _, peer := range a.peers {
		peers = append(peers, &Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].ID < peers[j].ID })
	return peers
}

func (a *ChatApp) isClosing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.closing
}

func (a *ChatApp) isSelfConnection(ip, port string) bool {
	if port != a.listenPort {
		return false
	}
	myIP := getMyIP()
	return ip == myIP || ip == "127.0.0.1" || ip == "localhost"
}

func validatePort(port string) (int, error) {
	p, err := strconv.Atoi(port)
	if err != nil {
		return 0, errors.New("port must be numeric")
	}
	if p < 1024 || p > 65535 {
		return 0, errors.New("port must be between 1024 and 65535")
	}
	return p, nil
}

func validateIPv4(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.To4() == nil {
		return errors.New("invalid IPv4 address")
	}
	return nil
}

func getMyIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			return localAddr.IP.String()
		}
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			return ip4.String()
		}
	}
	return "127.0.0.1"
}
