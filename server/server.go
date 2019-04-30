package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/tywkeene/telnet-chat/config"
	"github.com/tywkeene/telnet-chat/connection"
	"github.com/tywkeene/telnet-chat/room"
)

type Server struct {
	Running  bool
	Listener net.Listener
	Rooms    []*room.Room
	LogFile  *os.File
}

var helpMessage string = `
These commands are available while you are in a room:
/help: print this help message
/name: change your username
/leave: leave the current room and choose another
/quit: disconnect from the server
`

func (s *Server) ListRooms() string {
	str := "Available rooms:\n"
	for i, room := range s.Rooms {
		str += fmt.Sprintf("\t%d: %s\n", i, room.Name)
	}
	return str
}

func (s *Server) SelectRoom(c *connection.Connection) error {
	roomList := s.ListRooms()
	if err := c.SendMessage("Select a room to join by its number\n"); err != nil {
		return fmt.Errorf("Failed to send message to %q (%s): %s",
			c.UserName, c.Conn.RemoteAddr(), err.Error())
	}
	room, err := c.SendWithResponse(roomList)
	if err != nil {
		return fmt.Errorf("Failed to send message to %q (%s): %s\n",
			c.UserName, c.Conn.RemoteAddr(), err.Error())
	}

	if room == "" {
		c.SendError("Must enter room name to join")
		return fmt.Errorf("User %q (%s) failed to choose room\n", c.UserName, c.Conn.RemoteAddr())
	}

	roomIndex, err := strconv.Atoi(room)
	if err != nil {
		return fmt.Errorf("Error choosing room for user %s: %s\n", c.String(), err.Error())
	} else if roomIndex > len(s.Rooms) || roomIndex < 0 {
		return fmt.Errorf("User %s selected invalid room\n", c.String())
	}

	s.Rooms[roomIndex].AddUser(c)
	c.Room = roomIndex

	return nil
}

func (s *Server) HandleCommands(message string, c *connection.Connection) bool {
	switch message {
	case "/help":
		if err := c.SendMessage(helpMessage); err != nil {
			log.Println(err)
			return true
		}
	case "/name":
		newName, err := c.SendWithResponse("New name: ")
		if err != nil {
			log.Println(err)
			return true
		}
		log.Printf("User %s changed name to %s\n", c.String(), newName)
		s.Rooms[c.Room].WriteChan <- fmt.Sprintf("User %s changed name to %s\n", c.UserName, newName)
		c.UserName = newName

		return true
	case "/leave":
		room := s.Rooms[c.Room]
		room.RemoveUser(c)
		s.SelectRoom(c)
		return true
	case "/quit":
		c.Close()
		return true
	}
	return false
}

func (s *Server) HandleMessages(c *connection.Connection) {
	for c.Open == true {
		text, err := c.SendWithResponse(">> ")
		if err != nil {
			log.Printf("Failed to read message from %s: %s", c.String(), err.Error())
			return
		}

		if s.HandleCommands(text, c) == true {
			continue
		}

		message := fmt.Sprintf("<%s> (%s): %s\n", time.Now().Format(time.Kitchen), c.UserName, text)
		room := s.Rooms[c.Room]
		room.WriteMessage(message)

		logStr := fmt.Sprintf("%s: %s", room.Name, message)
		_, err = s.LogFile.WriteString(logStr)
		if err != nil {
			log.Printf("Failed to log message from user %s: %s", c.String(), err.Error())
		}

		log.Printf("User %s sent message %q to room %q\n", c.String(), text, room.Name)
	}
}

func (s *Server) HandleConnection(c *connection.Connection) {
	username, err := c.SendWithResponse("Desired username: ")
	if err != nil || username == "" {
		c.Close()
		log.Println("User failed to enter username")
		return
	}

	c.UserName = username
	log.Printf("User %s connected\n", c.String())

	if err := s.SelectRoom(c); err != nil {
		log.Println(err)
		c.Close()
		return
	}
	go s.HandleMessages(c)
}

func (s *Server) Serve() {

	for _, room := range s.Rooms {
		log.Printf("Starting room %q...\n", room.Name)
		go room.Run()
	}

	for s.Running {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		c := connection.NewConnection(conn)
		go s.HandleConnection(c)
	}
}

func (s *Server) InitializeRooms() {
	for _, roomName := range config.Config.Rooms {
		log.Printf("Initializing room %q\n", roomName)
		s.Rooms = append(s.Rooms, &room.Room{
			Name:        roomName,
			Connections: make(map[string]*connection.Connection, 0),
			WriteChan:   make(chan string),
		})
	}
}

func NewServer() (*Server, error) {

	bindAddr := config.Config.BindAddr + ":" + config.Config.BindPort

	log.Println("Starting listener on", bindAddr)
	listener, err := net.Listen("tcp4", bindAddr)
	if err != nil {
		return nil, err
	}

	log.Printf("Opening message log file %q\n", config.Config.LogFile)
	f, err := os.OpenFile(config.Config.LogFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Running:  true,
		Listener: listener,
		Rooms:    make([]*room.Room, 0),
		LogFile:  f,
	}

	s.InitializeRooms()

	return s, nil
}
