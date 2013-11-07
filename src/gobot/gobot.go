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

func Sample(arr []string) (string, error) {
  max := int64(len(arr))
  i, err := rand.Int(rand.Reader, big.NewInt(max))
  if err != nil {
    return "", err
  }
  return arr[i.Int64()], nil
}

var Plugins = map[string]func(con *irc.Connection, e irc.Event, replyName string){

  "/help|commands/": func(con *irc.Connection, e irc.Event, replyName string) {
    con.Privmsg(replyName, "roll, nextmeet, artme <string>, stab <nick>, seen <nick>, ram, uptime, 37status, boobs, trollface, dywj, dance, mustachify, stats, last, ping")
  },

  "meme": func(con *irc.Connection, e irc.Event, replyName string) {
    // There are no decent meme web services, nor gems wrapping the shitty ones.
    // -- Caius, 20th Aug 2011
    con.Privmsg(replyName, "Y U NO FIX MEME?!")
  },

  "/troll(face)?/": func(con *irc.Connection, e irc.Event, replyName string) {
    response, err := Sample([]string{"http://no.gd/troll.png", "http://no.gd/trolldance.gif", "http://caius.name/images/phone_troll.jpg"})
    if err != nil {
      return
    }

    con.Privmsg(replyName, response)
  },

	"version": func(con *irc.Connection, e irc.Event, replyName string) {
		reply := "My current version is"

		if GitCommit != "" {
			reply = fmt.Sprintf("%s %s", reply, GitCommit)
		} else {
			reply = fmt.Sprintf("%s unknown", reply)
		}

		if BuiltBy != "" {
			reply = fmt.Sprintf("%s and I was built by %s", reply, BuiltBy)
		}

		con.Privmsgf(replyName, reply)
	},

	// Pong plugin
	"/^(?:\\.|!?\\.?ping)$/": func(con *irc.Connection, e irc.Event, replyName string) {
		con.Privmsg(replyName, "pong!")
	},

	"/^stats?$/": func(con *irc.Connection, e irc.Event, replyName string) {
		con.Privmsg(replyName, "http://dev.hentan.caius.name/irc/nwrug.html")
	},

	"dance": func(con *irc.Connection, e irc.Event, replyName string) {
		i, err := rand.Int(rand.Reader, big.NewInt(3))
		if err != nil {
			i = big.NewInt(1)
		}

		switch i.Int64() {
		case 0:
			con.Privmsg(replyName, "EVERYBODY DANCE NOW!") // msg channel, "EVERYBODY DANCE NOW!"
			// TODO: ACTION
			con.Privmsg(replyName, "ACTION does the funky chicken")
		case 1:
			con.Privmsg(replyName, "http://no.gd/caiusboogie.gif")
		case 2:
			con.Privmsg(replyName, "http://i.imgur.com/rDDjz.gif")
		}
	},

	// Stabs what he is comanded to. Unless it's himself.
	// `stab blah` => `* gobot stabs blah`
	"/stab (.+)/": func(con *irc.Connection, e irc.Event, replyName string) {
		msg := e.Message

		stab_regexp := regexp.MustCompile("stab (.+)")

		receiver := stab_regexp.FindStringSubmatch(msg)[1]
		// If they try to stab us, stab them
		if strings.Contains(receiver, "rugbot") {
			receiver = e.Nick
		}

		// TODO: ACTION message
		con.Privmsgf(replyName, "/me stabs %s", receiver)
	},

	// Listens to channel conversation and inserts title of any link posted, following redirects
	// `And then I went to www.caius.name` => `gobot: Caius Durling &raquo; Profile`
	"http": func(con *irc.Connection, e irc.Event, replyName string) {
		msg := e.Message

		fmt.Printf("URLHandler checking '%s'\n", msg)

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
		con.Privmsg(replyName, title)
	},
}

func main() {
	fmt.Printf("Version: %s\nBuilt by: %s\n", GitCommit, BuiltBy) // FU GO

	roomName := "#caius"
	botName := "gobot"

	con := irc.IRC(botName, botName)
	err := con.Connect("irc.freenode.net:6667")
	if err != nil {
		log.Fatal("Can't connect to freenode")
	}

	con.AddCallback("001", func(e *irc.Event) {
		con.Join(roomName)
	})

	con.AddCallback("PRIVMSG", func(e *irc.Event) {
		// TODO: have irc.Event contain the room name for the PRIVMSG
		fmt.Printf("[%6s] %6s: %s\n", roomName, e.Nick, e.Message)

		// Plugins!
		lowerMessage := strings.ToLower(e.Message)

		for matchString := range Plugins {
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

			f := Plugins[matchString]
			go f(con, *e, roomName)
		}
	})

	con.Loop()
}
