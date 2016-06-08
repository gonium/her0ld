package main

import (
	"crypto/tls"
	"github.com/codegangsta/cli"
	"github.com/gonium/her0ld"
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
	app.Commands = []cli.Command{
		{
			Name:  "genconfig",
			Usage: "Generates a sample configuration file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: "her0ld.cfg",
					Usage: "the configuration file to use",
				},
			},
			Action: func(c *cli.Context) error {
				cfgfile := c.String("config")
				log.Printf("Writing sample configuration to %s", cfgfile)
				cfg := her0ld.MkExampleConfig()
				err := her0ld.SaveConfig(cfgfile, cfg)
				if err != nil {
					log.Fatalf("Failed to save config to file: %s", err.Error())
				}
				return nil
			},
		},
		{
			Name:  "run",
			Usage: "Run the bot",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: "her0ld.cfg",
					Usage: "the configuration file to use",
				},
				cli.BoolFlag{
					Name: "verbose",
					//Aliases: []string{"v"},
					Usage: "Print verbose output",
				},
			},
			Action: func(c *cli.Context) error {
				cfgfile := c.String("config")
				if c.Bool("verbose") {
					log.Printf("Started %s with verbose logging", app.Name)
					log.Printf("Using configuration found in %s", cfgfile)
				}
				cfg, err := her0ld.LoadConfig(cfgfile)
				if err != nil {
					log.Fatalf("Failed to load config: %s", err.Error())
				}

				// TODO: Deal with several channel/nick configurations
				channel := cfg.Bots[0].Channel
				nick := cfg.Bots[0].Nick
				fullname := cfg.Bots[0].Fullname
				server := cfg.Bots[0].Server
				quitmsg := cfg.Bots[0].Quitmsg
				enable_tls := cfg.Bots[0].EnableTLS
				strictcertcheck := cfg.Bots[0].StrictCertCheck

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

				ircconn := irc.IRC(nick, fullname)
				if c.Bool("verbose") {
					ircconn.Debug = true
					//ircconn.VerboseCallbackHandler = true
				}
				ircconn.UseTLS = enable_tls
				ircconn.TLSConfig = &tls.Config{
					InsecureSkipVerify: !strictcertcheck,
				}
				ircconn.PingFreq = 1 * time.Minute
				ircconn.QuitMessage = quitmsg

				var allBots []her0ldbot.Bot
				if cfg.Functions.Echobot_enable {
					// the echobot is only helpful during development...
					allBots = append(allBots, her0ldbot.NewEchoBot("Echobot"))
				}
				if cfg.Functions.Pingbot_enable {
					allBots = append(allBots, her0ldbot.NewPingBot("Pingbot"))
				}
				if cfg.Functions.Eventbot_enable {
					// TODO: Move settings to the configuration file
					allBots = append(allBots, her0ldbot.NewEventBot("Eventbot",
						cfg.EventbotCfg))
				}

				// Join channel upon welcome message
				ircconn.AddCallback("001", func(e *irc.Event) {
					ircconn.Join(channel)
				})
				// When end of nick list of channel received: send hello message
				// to channel
				ircconn.AddCallback("366", func(e *irc.Event) {
					//ircconn.Privmsg(channel, "bot is active")
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
					msg := her0ldbot.InboundMessage{
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
									currentnick := ircconn.GetNick()
									if line.Destination != currentnick {
										ircconn.Privmsg(line.Destination, line.Message)
									} else {
										log.Printf("Line destination is %s, would send to myself (%s). Dropping line.", line.Destination, currentnick)
									}

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
									currentnick := ircconn.GetNick()
									if line.Destination != currentnick {
										ircconn.Privmsg(line.Destination, line.Message)
									} else {
										log.Printf("Line destination is %s, would send to myself (%s). Dropping line.", line.Destination, currentnick)
									}
								}
							}
						}

					}
				})

				// now, run the server
				err = ircconn.Connect(server)
				if err != nil {
					log.Fatalf("Failed to connect to %s: %s", server, err.Error())
				}

				// wait for termination signal
				<-quit
				// cleanup tasks
				log.Printf("Terminating bot.")
				ircconn.Quit()
				time.Sleep(1 * time.Second)

				return nil
			},
		},
	}
	app.Run(os.Args)
}
