package main

import (
	"crypto/rand"
	"fmt"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
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

func main() {
	fmt.Printf("Version: %s\nBuilt by: %s\n", GitCommit, BuiltBy) // FU GO

	bot := Gobot{Name: "gobot", Room: "#caius", Server: "irc.freenode.net:6667"}
	bot.Plugins = make(map[string]func(p Privmsg))

	bot.Match("hullo", func(privmsg Privmsg) {
		privmsg.Msg("Oh hai!")
	})

	bot.Match("/help|commands/", func(privmsg Privmsg) {
		privmsg.Msg("roll, nextmeet, artme <string>, stab <nick>, seen <nick>, ram, uptime, 37status, boobs, trollface, dywj, dance, mustachify, stats, last, ping")
	})

	bot.Match("meme", func(privmsg Privmsg) {
		// There are no decent meme web services, nor gems wrapping the shitty ones.
		// -- Caius, 20th Aug 2011
		privmsg.Msg("Y U NO FIX MEME?!")
	})

	bot.Match("/troll(face)?/", func(privmsg Privmsg) {
		response, err := bot.Sample([]string{"http://no.gd/troll.png", "http://no.gd/trolldance.gif", "http://caius.name/images/phone_troll.jpg"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("boner", func(privmsg Privmsg) {
		response, err := bot.Sample([]string{"http://files.myopera.com/coxy/albums/106123/trex-boner.jpg", "http://no.gd/badger.gif"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("badger", func(privmsg Privmsg) {
		privmsg.Msg("http://no.gd/badger2.gif")
	})

	bot.Match("dywj", func(privmsg Privmsg) {
		privmsg.Msg("DAMN YOU WILL JESSOP!!!")
	})

	// derp, herp
	bot.Match("/\\b[dh]erp\\b/", func(privmsg Privmsg) {
		privmsg.Msg("http://caius.name/images/qs/herped-a-derp.png")
	})

	bot.Match("/F{2,}U{2,}/", func(privmsg Privmsg) {
		var response string

		if strings.Contains(strings.ToLower(privmsg.Nick), "tomb") {
			response = "http://no.gd/p/calm-20111107-115310.jpg"
		} else {
			response = fmt.Sprintf("Calm down %s!", privmsg.Nick)
		}

		privmsg.Msg(response)
	})

	bot.Match("nextmeat", func(privmsg Privmsg) {
		privmsg.Msg("BACNOM")
	})

	bot.Match("/where is (wlll|will)/", func(privmsg Privmsg) {
		response, err := bot.Sample([]string{"North Tea Power", "home"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("/^b(oo|ew)bs$/", func(privmsg Privmsg) {
		response, err := bot.Sample([]string{"(.)(.)", "http://no.gd/boobs.gif"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("version", func(privmsg Privmsg) {
		reply := "My current version is"

		if GitCommit != "" {
			reply = fmt.Sprintf("%s %s", reply, GitCommit)
		} else {
			reply = fmt.Sprintf("%s unknown", reply)
		}

		if BuiltBy != "" {
			reply = fmt.Sprintf("%s and I was built by %s", reply, BuiltBy)
		}

		privmsg.Msg(reply)
	})

	// Pong plugin
	bot.Match("/^(?:\\.|!?\\.?ping)$/", func(privmsg Privmsg) {
		privmsg.Msg("pong!")
	})

	bot.Match("/^stats?$/", func(privmsg Privmsg) {
		privmsg.Msg("http://dev.hentan.caius.name/irc/nwrug.html")
	})

	bot.Match("dance", func(privmsg Privmsg) {
		i, err := rand.Int(rand.Reader, big.NewInt(3))
		if err != nil {
			i = big.NewInt(1)
		}

		switch i.Int64() {
		case 0:
			privmsg.Msg("EVERYBODY DANCE NOW!") // msg channel, "EVERYBODY DANCE NOW!"
			privmsg.Action("does the funky chicken")
		case 1:
			privmsg.Msg("http://no.gd/caiusboogie.gif")
		case 2:
			privmsg.Msg("http://i.imgur.com/rDDjz.gif")
		}
	})

	// Stabs what he is comanded to. Unless it's himself.
	// `stab blah` => `* gobot stabs blah`
	bot.Match("/stab (.+)/", func(privmsg Privmsg) {
		msg := privmsg.Message

		stab_regexp := regexp.MustCompile("stab (.+)")

		receiver := stab_regexp.FindStringSubmatch(msg)[1]
		// If they try to stab us, stab them
		if strings.Contains(receiver, "rugbot") {
			receiver = privmsg.Nick
		}

		// TODO: privmsg.Actionf()
		privmsg.Action(fmt.Sprintf("/me stabs %s", receiver))
	})

	// Listens to channel conversation and inserts title of any link posted, following redirects
	// `And then I went to www.caius.name` => `gobot: Caius Durling &raquo; Profile`
	bot.Match("/.+/", func(privmsg Privmsg) {
		msg := privmsg.Message

		// Regexp from http://daringfireball.net/2010/07/improved_regex_for_matching_urls - Ta gruber!
		url_regexp := regexp.MustCompile("(?i)\\b((?:https?://|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\\s()<>]+|\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\))+(?:\\(([^\\s()<>]+|(\\([^\\s()<>]+\\)))*\\)|[^\\s`!()\\[\\]{};:'\".,<>?«»“”‘’]))")
		url := url_regexp.FindString(msg)

		if url == "" {
			return
		}

		fmt.Printf("Extracted '%s'\n", url)

		// We might extract `www.google.com` or `bit.ly/something` so we need to prepend http:// in that case
		if !strings.HasPrefix(url, "http://") {
			url = fmt.Sprintf("http://%s", url)
		}

		fmt.Printf("GET %s\n", url)

		// Attempt a GET request to get the page title
		// TODO: handle PDF and non-HTML content
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		raw_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		body := string(raw_body)

		title_regexp := regexp.MustCompile("<title>([^<]+)</title>")
		title := title_regexp.FindStringSubmatch(body)[1]

		fmt.Printf("title: %s\n", title)

		privmsg.Msg(title)
	})

	// TODO: last
	// TODO: roll
	// TODO: ACTION pokes .+
	// TODO: 37status
	// TODO: hubstatus
	// TODO: nextmeet
	// TODO: ACTION staabs
	// TODO: artme
	// TODO: tasche http
	// TODO: tasche artme
	// TODO: seen
	// TODO: ram
	// TODO: uptime
	// TODO: last poop
	// TODO: twitter status
	// TODO: twitter user
	// TODO: commit me

	bot.Run()
}
