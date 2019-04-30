package room

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/tywkeene/telnet-chat/connection"
)

type Room struct {
	sync.RWMutex
	Name        string
	Connections map[string]*connection.Connection
	WriteChan   chan string
}

func (r *Room) RemoveUser(c *connection.Connection) {
	r.Lock()
	log.Printf("Removing user %s from room %s\n", c.UserName, r.Name)
	r.WriteChan <- fmt.Sprintf("<%s> User %s left.\n", r.Name, c.UserName)
	delete(r.Connections, c.UserName)
	r.Unlock()
}

func (r *Room) AddUser(c *connection.Connection) {
	r.Lock()
	log.Printf("Adding user %s to room %s\n", c.String(), r.Name)
	r.Connections[c.UserName] = c
	r.WriteChan <- fmt.Sprintf("<%s> User %s joined.\n", r.Name, c.UserName)

	welcomeMessage := fmt.Sprintf("Welcome to %q!\n", r.Name)
	c.SendMessage(welcomeMessage)
	r.Unlock()
}

func (r *Room) Run() {
	for {
		message := <-r.WriteChan
		for _, c := range r.Connections {
			if !strings.Contains(message, c.UserName) {
				if err := c.SendMessage(message); err != nil {
					log.Printf("Failed to write message in room %q to user %s: %s\n",
						r.Name, c.String(), err.Error())
				}
			}
		}
	}
}
