package main

import (
  "github.com/caius/gobot"
)

func main() {
  // Create it!
  bot := gobot.Gobot{Name: "ponger", Room: "#caius", Server: "irc.freenode.net:6667"}
	bot.Plugins = make(map[string]func(p gobot.Privmsg)) // Ew

  // When someone says ping, respond with pong!
  bot.Match("ping", func(p gobot.Privmsg) {
    p.Msg("Pong!!")
  })

  // Connect to server, join room, listen for messages
  bot.Run()
}
