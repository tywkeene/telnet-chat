package room

import (
	"fmt"
	"log"
	"strings"

	"github.com/tywkeene/telnet-chat/connection"
)

type Room struct {
	Name        string
	Connections []*connection.Connection
	WriteChan   chan string
}

func (r *Room) AddUser(c *connection.Connection) {
	log.Printf("Adding user %s to room %s\n", c.String(), r.Name)
	r.Connections = append(r.Connections, c)
	r.WriteChan <- fmt.Sprintf("<%s> User %s joined.\n", r.Name, c.UserName)

	welcomeMessage := fmt.Sprintf("Welcome to %q!\n", r.Name)
	c.SendMessage(welcomeMessage)
}

func (r *Room) Run() {
	for {
		message := <-r.WriteChan
		for _, conn := range r.Connections {
			if !strings.Contains(message, conn.UserName) {
				if err := conn.SendMessage(message); err != nil {
					log.Printf("Failed to write message in room %q to user %s: %s\n",
						r.Name, conn.String(), err.Error())
				}
			}
		}
	}
}
