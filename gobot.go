package gobot

import (
	"crypto/rand"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"math/big"
	"regexp"
)

type Bot struct {
	Name string // Bot nick
	Pass string // Password for nickserv, if required (noop if empty)
	// TODO: make an array once PRIVMSG handler can work out source of event
	Room   string
	Server string // "irc.freenode.net"
	Port   int    // 6667

	Plugins map[*regexp.Regexp]func(p Privmsg)

	Con *irc.Connection
}

func Gobot() Bot {
	gobot := Bot{}
	gobot.Plugins = make(map[*regexp.Regexp]func(p Privmsg))
	return gobot
}

func (bot *Bot) Address() string {
	return fmt.Sprintf("%s:%d", bot.Server, bot.Port)
}

// Run the bot. Setup plugins before calling this, blocks execution until program end
func (bot *Bot) Run() {
	bot.Con = irc.IRC(bot.Name, bot.Name) // Use the bot's nick as real name too
	err := bot.Con.Connect(bot.Address())
	if err != nil {
		log.Fatal("Couldn't connect to %s: %s", bot.Address, err)
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
		for matcher := range bot.Plugins {
			if !matcher.MatchString(e.Message) {
				fmt.Printf("Skipping %s plugin\n", matcher)
				continue
			}

			f := bot.Plugins[matcher]
			f(privmsg)
		}
	})

	bot.Con.Loop()
}

func (bot *Bot) NickRegexp(input string) *regexp.Regexp {
	// TODO: case insensitive
	input = fmt.Sprintf("(?:\\A|%s:? )%s\\z", bot.Name, input)
	return regexp.MustCompile(input)
}

// Add handler for string trigger, using nick regexp to wrap the string
func (bot *Bot) MatchString(trigger string, handler func(privmsg Privmsg)) {
	// Match full phrase, optionally prepended with <nick>: or <nick>
	// eg:
	//    `help'
	//    `mybot: help'
	//    `mybot help'
	matcher := bot.NickRegexp(trigger)
	bot.Match(matcher, handler)
}

// Add handler for regexp
// regexp is untouched
func (bot *Bot) Match(trigger *regexp.Regexp, handler func(privmsg Privmsg)) {
	bot.Plugins[trigger] = handler
}

// Helper method for returning "random" responses
func (bot *Bot) Sample(arr []string) (string, error) {
	max := int64(len(arr))
	i, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	return arr[i.Int64()], nil
}
