package gobot

import (
  "fmt"
	"github.com/thoj/go-ircevent"
)

type Privmsg struct {
	Event      irc.Event
	Message    string
	Nick       string
	Connection irc.Connection
	RoomName   string
}

func (self *Privmsg) Msg(response string) {
	self.Connection.Privmsg(self.RoomName, response)
}

func (self *Privmsg) Action(response string) {
	// TODO: implement ACTION
	fmt.Println("TODO: implement ACTION (dickhead)")
	self.Msg(response)
}

func (self *Privmsg) Error(response error) {
  self.Connection.Privmsgf(self.RoomName, "OMG ERROR!!!1111!!11o1oneoneone - %s", response)
}
