package main

import (
  "log"
  "fmt"
  "github.com/thoj/go-ircevent"
  "strings"
  "regexp"
  "net/http"
  "io/ioutil"
)


// Listens to channel conversation and inserts title of any link posted, following redirects
// `And then I went to www.caius.name` => `gobot: Caius Durling &raquo; Profile`
func URLHandler(con *irc.Connection, e irc.Event, replyName string) {
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
}

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
    // TODO: have irc.Event contain the room name for the PRIVMSG
    fmt.Printf("[%6s] %6s: %s\n", roomName, e.Nick, e.Message)

    // "Plugins"
    URLHandler(con, *e, roomName)
  })

  con.Loop()
}
