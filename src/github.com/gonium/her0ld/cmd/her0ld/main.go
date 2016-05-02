package main

import (
	"crypto/tls"
	"github.com/codegangsta/cli"
	"github.com/gonium/her0ld/bots"
	irc "github.com/thoj/go-ircevent"
	"log"
	"os"
	"os/signal"
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
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from given file",
		},
	}
	// TODO: Create a command for generating an example configuration
	app.Commands = []cli.Command{
		{
			Name:  "genconfig",
			Usage: "Generates a sample configuration file",
			Action: func(c *cli.Context) {
				cfgfile := c.String("config")
				if cfgfile == "" {
					log.Fatalf("Please provide a configuration file (-c)")
				}
				if c.Bool("verbose") {
					log.Println("Writing sample configuration to %s", cfgfile)
				}
			},
		},
	}
	app.Action = func(c *cli.Context) {
		channel := "#her0ld-dev"
		nick := "her0ld-dev"
		server := "irc.hackint.org:9999"
		quitmsg := `There's been a little complication with my complication.
										-- Mrs. Terrain, Brazil`

		// whenever the bot should be terminated, send a 'true' to this
		// channel.
		quit := make(chan bool)

		// intercept CTRL-C and jump to proper bot termination
		ctrlc := make(chan os.Signal, 1)
		signal.Notify(ctrlc, os.Interrupt)
		go func() {
			for _ = range ctrlc {
				if c.Bool("verbose") {
					log.Printf("Received SIGINT")
				}
				quit <- true
			}
		}()

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

		var allBots []her0ld.Bot
		// the echobot is only helpful during development...
		allBots = append(allBots, her0ld.NewEchoBot("Echobot"))
		allBots = append(allBots, her0ld.NewPingBot("Pingbot"))

		// Join channel upon welcome message
		ircconn.AddCallback("001", func(e *irc.Event) {
			ircconn.Join(channel)
		})
		// When end of nick list of channel received: send hello message
		// to channel
		ircconn.AddCallback("366", func(e *irc.Event) {
			ircconn.Privmsg(channel, "bot is active")
		})

		// forward PRIVMSG to bots
		ircconn.AddCallback("PRIVMSG", func(event *irc.Event) {
			//event.Message() contains the message
			//event.Nick Contains the sender
			//event.Arguments[0] Contains the channel
			for idx, arg := range event.Arguments {
				log.Println("%d - %s", idx, arg)
			}
			source := event.Arguments[0]
			msg := her0ld.InboundMessage{
				Channel: source,
				Nick:    event.Nick,
				Message: event.Message(),
			}
			if msg.IsChannelEvent() {
				// channel message
				log.Printf("Inbound channel message: %s", msg)
				for _, bot := range allBots {
					answerlines, err := bot.ProcessChannelEvent(msg)
					if err != nil {
						log.Printf("Bot %s failed to process inbound channel message \"%s\": %s",
							bot.GetName(), event.Message(), err.Error())
					} else {
						for _, line := range answerlines {
							ircconn.Privmsg(line.Destination, line.Message)
						}
					}
				}
			} else {
				// query message
				// TODO: This is broken. event.Nick does not contain the sender,
				// but the bot itself.
				log.Printf("Inbound query message: %s", msg)
				for _, bot := range allBots {
					answerlines, err := bot.ProcessQueryEvent(msg)
					if err != nil {
						log.Printf("Bot %s failed to process inbound query message \"%s\": %s",
							bot.GetName(), event.Message(), err.Error())
					} else {
						for _, line := range answerlines {
							ircconn.Privmsg(line.Destination, line.Message)
						}
					}
				}

			}
		})

		// now, run the server
		err := ircconn.Connect(server)
		if err != nil {
			log.Fatalf("Failed to connect to %s: %s", server, err.Error())
		}

		// wait for termination signal
		<-quit
		// cleanup tasks
		log.Printf("Terminating bot.")
		ircconn.Quit()
		time.Sleep(1 * time.Second)
	}
	app.Run(os.Args)
}
