package main

import (
  "log"
  "fmt"
  "github.com/thoj/go-ircevent"
)

func main() {
  fmt.Printf("") // FU GO

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
    fmt.Printf("[%6s] %6s: %s\n", e.Nick, e.Message)

    // "Plugins" go here
  })

  con.Loop()
}
