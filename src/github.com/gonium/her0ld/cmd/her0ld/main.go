package main

import (
	//"crypto/tls"
	"github.com/codegangsta/cli"
	irc "github.com/fluffle/goirc/client"
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
		if c.Bool("verbose") {
			log.Printf("Started %s", app.Name)
		}

		cfg := irc.NewConfig("her0ld-dev")
		cfg.SSL = false
		//cfg.SSLConfig = &tls.Config{ServerName: "irc.hackint.org"}
		//cfg.Server = "irc.hackint.org:9999"
		cfg.Server = "irc.hackint.org:6667"
		cfg.NewNick = func(n string) string { return n + "^" }
		ircconn := irc.Client(cfg)
		ircconn.EnableStateTracking()

		// Add handlers to do things here!
		// e.g. join a channel on connect.
		ircconn.HandleFunc(irc.CONNECTED,
			func(conn *irc.Conn, line *irc.Line) {
				log.Printf("Attempting to join channel")
				conn.Join("#her0ld-dev")
			})
		// And a signal on disconnect
		quit := make(chan bool)
		ircconn.HandleFunc(irc.DISCONNECTED,
			func(conn *irc.Conn, line *irc.Line) {
				log.Printf("Received DISCONNECTED: \n- %#v  \n-%#v", conn, line)
				quit <- true
			})

		// Tell client to connect.
		if c.Bool("verbose") {
			log.Printf("Attempting to connect to %s", cfg.Server)
		}
		if err := ircconn.Connect(); err != nil {
			log.Fatal("Connection error: %s\n", err.Error())
		}

		// TODO: Print all lines received from server.
		go func() {
			for fromserver := range ircconn.in {
				log.Println(fromserver)
			}
		}()

		log.Println("Client: %#v", ircconn)
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			ircconn.Privmsg("#her0ld-dev", "foobar!")
			ircconn.Notice("#her0ld-dev", "notice!")
		}

		// Wait for disconnect
		<-quit

	}
	app.Run(os.Args)
}
