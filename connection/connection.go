package connection

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"syscall"
)

type Connection struct {
	Conn     net.Conn
	Buffer   *bufio.Reader
	UserName string
	Room     int
	Open     bool
}

func (c *Connection) String() string {
	return fmt.Sprintf("%q (%s)", c.UserName, c.Conn.RemoteAddr())
}

func (c *Connection) Close() {
	log.Printf("User %s disconnected\n", c.String())
	c.SendMessage("Goodbye!\n")
	c.Open = false
	c.Conn.Close()
}

func (c *Connection) SendMessage(str string) error {
	_, err := c.Conn.Write([]byte(str))
	if err == io.EOF || err == syscall.EINVAL {
		c.Conn.Close()
		return fmt.Errorf("Client closed connection")
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) ReadMessage() (string, error) {
	response, _, err := c.Buffer.ReadLine()
	if err == io.EOF || err == syscall.EINVAL {
		c.Close()
		return "", fmt.Errorf("Client closed connection")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(response)), nil
}

func (c *Connection) SendWithResponse(message string) (string, error) {
	if err := c.SendMessage(message); err != nil {
		return "", err
	}
	return c.ReadMessage()
}

func (c *Connection) SendError(str string) error {
	log.Printf("Sending error to %s: %s", c.Conn.RemoteAddr(), str)
	return c.SendMessage("Server error: " + str)
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn:   conn,
		Buffer: bufio.NewReader(conn),
		Open:   true,
	}
}
