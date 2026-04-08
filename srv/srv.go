package srv

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type PeerConn struct {
	Conn       net.Conn
	Reader     *bufio.Reader
	RemoteIP   string
	RemotePort string
}

type Handlers struct {
	IsClosing        func() bool
	ValidateIncoming func(ip, port string) error
	OnConnect        func(*PeerConn)
	OnMessage        func(*PeerConn, string)
	OnDisconnect     func(*PeerConn, bool)
	OnAcceptError    func(error)
}

func Start(port string) (net.Listener, error) {
	return net.Listen("tcp", ":"+port)
}

func AcceptLoop(listener net.Listener, myListenPort string, handlers Handlers) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if handlers.IsClosing != nil && handlers.IsClosing() {
				return
			}
			if handlers.OnAcceptError != nil {
				handlers.OnAcceptError(err)
			}
			continue
		}
		go handleIncoming(conn, myListenPort, handlers)
	}
}

func handleIncoming(conn net.Conn, myListenPort string, handlers Handlers) {
	reader := bufio.NewReader(conn)

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := reader.ReadString('\n')
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		_ = conn.Close()
		if handlers.OnAcceptError != nil {
			handlers.OnAcceptError(fmt.Errorf("failed to complete incoming handshake from %s: %w", conn.RemoteAddr(), err))
		}
		return
	}

	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) != 2 || parts[0] != "HELLO" {
		_, _ = fmt.Fprintln(conn, "ERROR invalid handshake")
		_ = conn.Close()
		return
	}

	remotePort := parts[1]
	remoteIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		remoteIP = conn.RemoteAddr().String()
	}

	if handlers.ValidateIncoming != nil {
		if err := handlers.ValidateIncoming(remoteIP, remotePort); err != nil {
			_, _ = fmt.Fprintln(conn, "ERROR "+err.Error())
			_ = conn.Close()
			return
		}
	}

	if _, err := fmt.Fprintf(conn, "HELLO_ACK %s\n", myListenPort); err != nil {
		_ = conn.Close()
		return
	}

	peer := &PeerConn{
		Conn:       conn,
		Reader:     reader,
		RemoteIP:   remoteIP,
		RemotePort: remotePort,
	}

	if handlers.OnConnect != nil {
		handlers.OnConnect(peer)
	}

	MonitorPeer(peer, handlers)
}

func MonitorPeer(peer *PeerConn, handlers Handlers) {
	for {
		line, err := peer.Reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				_ = peer.Conn.Close()
			}
			if handlers.OnDisconnect != nil {
				handlers.OnDisconnect(peer, false)
			}
			return
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "MSG "):
			if handlers.OnMessage != nil {
				handlers.OnMessage(peer, strings.TrimPrefix(line, "MSG "))
			}
		case line == "BYE":
			if handlers.OnDisconnect != nil {
				handlers.OnDisconnect(peer, true)
			}
			return
		default:
			if handlers.OnMessage != nil {
				handlers.OnMessage(peer, "")
			}
		}
	}
}
