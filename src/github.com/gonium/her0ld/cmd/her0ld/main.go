package main

import (
	"crypto/tls"
	"github.com/codegangsta/cli"
	irc "github.com/thoj/go-ircevent"
	"log"
	"os"
	"time"
)

func main() {
	app := cli.NewApp()
	app.Name = "her0ld"
	app.Usage = "A friendly IRC bot written in Go."
	app.Version = "0.1.0"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "print verbose messages",
		},
	}
	app.Action = func(c *cli.Context) {
		channel := "#her0ld-dev"
		nick := "her0ld-dev"
		server := "irc.hackint.org:9999"
		quitmsg := "There's been a little complication with my complication. -- Mr. Terrain, Brazil"

		// TODO: proper event handling
		quit := make(chan bool)

		ircconn := irc.IRC(nick, "Development of her0ld bot")
		if c.Bool("verbose") {
			log.Printf("Started %s with verbose logging", app.Name)
			ircconn.Debug = true
			//ircconn.VerboseCallbackHandler = true
		}
		ircconn.UseTLS = true
		ircconn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		ircconn.PingFreq = 1 * time.Minute
		ircconn.QuitMessage = quitmsg

		// Join channel upon welcome message
		ircconn.AddCallback("001", func(e *irc.Event) {
			ircconn.Join(channel)
		})
		// When end of nick list of channel received: send hello message
		// to channel
		ircconn.AddCallback("366", func(e *irc.Event) {
			ircconn.Privmsg(channel, "bot is active")
		})
		ircconn.Connect(server)

		//// Wait for disconnect
		<-quit

	}
	app.Run(os.Args)
}
