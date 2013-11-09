package gobot

import (
	"crypto/rand"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"math/big"
	"regexp"
	"strings"
)

var GitCommit string
var BuiltBy string

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

type Gobot struct {
	Name   string // Bot nick
	Pass   string // Password for nickserv, if required (noop if empty)
	Room   string // TODO: make an array once PRIVMSG handler can work out source of event
	Server string // "server.name:port"

	Plugins map[string]func(p Privmsg)

	Con *irc.Connection
}

// Run the bot. Setup plugins before calling this, blocks execution until program end
func (bot *Gobot) Run() {
	bot.Con = irc.IRC(bot.Name, bot.Name) // Use the bot's nick as real name too
	err := bot.Con.Connect(bot.Server)
	if err != nil {
		log.Fatal("Couldn't connect to %s: %s", bot.Server, err)
	}

	// Once we're successfully connected to the network
	bot.Con.AddCallback("001", func(e *irc.Event) {
		// Join our rooms!
		bot.Con.Join(bot.Room)
	})

	// Handle messages from rooms we're in
	bot.Con.AddCallback("PRIVMSG", func(e *irc.Event) {
		// TODO: have irc.Event contain the room name for the PRIVMSG
		fmt.Printf("[%6s] %6s: %s\n", bot.Name, e.Nick, e.Message)

		privmsg := Privmsg{Connection: *bot.Con, Event: *e, Message: e.Message, Nick: e.Nick, RoomName: bot.Room}

		// Plugins!
		lowerMessage := strings.ToLower(privmsg.Message)

		for matchString := range bot.Plugins {
			// TODO: check matchString against e.Message
			if strings.HasPrefix(matchString, "/") && strings.HasPrefix(matchString, "/") {
				// Yer a regexp Harry
				regexString := strings.TrimPrefix(matchString, "/")
				regexString = strings.TrimSuffix(regexString, "/")

				fmt.Printf("regexString: %s\n", regexString)

				// Get on yer bike if we don't match
				if !regexp.MustCompile(regexString).MatchString(e.Message) {
					fmt.Printf("Skipping %s plugin - no regexp match\n", matchString)
					continue
				}
			} else {
				// Just a plain ole string
				if !strings.Contains(lowerMessage, matchString) {
					fmt.Printf("Skipping %s plugin - no string match\n", matchString)
					continue
				}
			}

			f := bot.Plugins[matchString]
			f(privmsg)
		}
	})

	bot.Con.Loop()
}

// Add handler for trigger
func (bot *Gobot) Match(trigger string, handler func(privmsg Privmsg)) {
	bot.Plugins[trigger] = handler
}

// Helper method for returning "random" responses
func (bot *Gobot) Sample(arr []string) (string, error) {
	max := int64(len(arr))
	i, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	return arr[i.Int64()], nil
}
