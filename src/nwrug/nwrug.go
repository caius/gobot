package main

import (
	"fmt"
	"gobot"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func main() {
	fmt.Printf("Version: %s\nBuilt by: %s\n", gobot.GitCommit, gobot.BuiltBy) // FU GO

	bot := gobot.Gobot{Name: "gobot", Room: "#caius", Server: "irc.freenode.net:6667"}
	bot.Plugins = make(map[string]func(p gobot.Privmsg))

	bot.Match("hullo", func(privmsg gobot.Privmsg) {
		privmsg.Msg("Oh hai!")
	})

	bot.Match("/help|commands/", func(privmsg gobot.Privmsg) {
		privmsg.Msg("roll, nextmeet, artme <string>, stab <nick>, seen <nick>, ram, uptime, 37status, boobs, trollface, dywj, dance, mustachify, stats, last, ping")
	})

	bot.Match("meme", func(privmsg gobot.Privmsg) {
		// There are no decent meme web services, nor gems wrapping the shitty ones.
		// -- Caius, 20th Aug 2011
		privmsg.Msg("Y U NO FIX MEME?!")
	})

	bot.Match("/troll(face)?/", func(privmsg gobot.Privmsg) {
		response, err := bot.Sample([]string{"http://no.gd/troll.png", "http://no.gd/trolldance.gif", "http://caius.name/images/phone_troll.jpg"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("boner", func(privmsg gobot.Privmsg) {
		response, err := bot.Sample([]string{"http://files.myopera.com/coxy/albums/106123/trex-boner.jpg", "http://no.gd/badger.gif"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("badger", func(privmsg gobot.Privmsg) {
		privmsg.Msg("http://no.gd/badger2.gif")
	})

	bot.Match("dywj", func(privmsg gobot.Privmsg) {
		privmsg.Msg("DAMN YOU WILL JESSOP!!!")
	})

	// derp, herp
	bot.Match("/\\b[dh]erp\\b/", func(privmsg gobot.Privmsg) {
		privmsg.Msg("http://caius.name/images/qs/herped-a-derp.png")
	})

	bot.Match("/F{2,}U{2,}/", func(privmsg gobot.Privmsg) {
		var response string

		if strings.Contains(strings.ToLower(privmsg.Nick), "tomb") {
			response = "http://no.gd/p/calm-20111107-115310.jpg"
		} else {
			response = fmt.Sprintf("Calm down %s!", privmsg.Nick)
		}

		privmsg.Msg(response)
	})

	bot.Match("nextmeat", func(privmsg gobot.Privmsg) {
		privmsg.Msg("BACNOM")
	})

	bot.Match("/where is (wlll|will)/", func(privmsg gobot.Privmsg) {
		response, err := bot.Sample([]string{"North Tea Power", "home"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("/^b(oo|ew)bs$/", func(privmsg gobot.Privmsg) {
		response, err := bot.Sample([]string{"(.)(.)", "http://no.gd/boobs.gif"})
		if err != nil {
			return
		}

		privmsg.Msg(response)
	})

	bot.Match("version", func(privmsg gobot.Privmsg) {
		reply := "My current version is"

		if gobot.GitCommit != "" {
			reply = fmt.Sprintf("%s %s", reply, gobot.GitCommit)
		} else {
			reply = fmt.Sprintf("%s unknown", reply)
		}

		if gobot.BuiltBy != "" {
			reply = fmt.Sprintf("%s and I was built by %s", reply, gobot.BuiltBy)
		}

		privmsg.Msg(reply)
	})

	// Pong plugin
	bot.Match("/^(?:\\.|!?\\.?ping)$/", func(privmsg gobot.Privmsg) {
		privmsg.Msg("pong!")
	})

	bot.Match("/^stats?$/", func(privmsg gobot.Privmsg) {
		privmsg.Msg("http://dev.hentan.caius.name/irc/nwrug.html")
	})

	bot.Match("dance", func(privmsg gobot.Privmsg) {
		i, err := bot.Sample([]string{"0", "1", "2"})
		if err != nil {
			return
		}

		switch i {
		case "0":
			privmsg.Msg("EVERYBODY DANCE NOW!") // msg channel, "EVERYBODY DANCE NOW!"
			privmsg.Action("does the funky chicken")
		case "1":
			privmsg.Msg("http://no.gd/caiusboogie.gif")
		case "2":
			privmsg.Msg("http://i.imgur.com/rDDjz.gif")
		}
	})

	// Stabs what he is comanded to. Unless it's himself.
	// `stab blah` => `* gobot stabs blah`
	bot.Match("/stab (.+)/", func(privmsg gobot.Privmsg) {
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
	bot.Match("/.+/", func(privmsg gobot.Privmsg) {
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
