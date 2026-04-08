package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type Session struct {
	Conn       net.Conn
	Reader     *bufio.Reader
	RemoteIP   string
	RemotePort string
}

func Start(ip, port, myListenPort string) (*Session, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(ip, port))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	if err := sendLine(conn, "HELLO "+myListenPort); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to send handshake: %w", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := reader.ReadString('\n')
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}

	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "ERROR ") {
		_ = conn.Close()
		return nil, errors.New(strings.TrimPrefix(line, "ERROR "))
	}

	parts := strings.Fields(line)
	if len(parts) != 2 || parts[0] != "HELLO_ACK" {
		_ = conn.Close()
		return nil, errors.New("invalid handshake response")
	}
	if parts[1] != port {
		_ = conn.Close()
		return nil, errors.New("peer responded with mismatched listening port")
	}

	return &Session{
		Conn:       conn,
		Reader:     reader,
		RemoteIP:   ip,
		RemotePort: port,
	}, nil
}

func SendMessage(conn net.Conn, msg string) error {
	return sendLine(conn, "MSG "+msg)
}

func SendTerminate(conn net.Conn) error {
	return sendLine(conn, "BYE")
}

func sendLine(conn net.Conn, line string) error {
	_, err := fmt.Fprintln(conn, line)
	return err
}
