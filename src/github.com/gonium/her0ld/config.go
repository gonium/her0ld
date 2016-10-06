package her0ld

import (
	"github.com/BurntSushi/toml"
	"os"
)

type BotConnection struct {
	Enabled         bool
	Channel         string
	Server          string
	Nick            string
	Fullname        string
	Quitmsg         string
	EnableTLS       bool
	StrictCertCheck bool
}

type GeneralConfig struct {
	OwnerNick        string
	OwnerEmailAdress string
}

type EmailSettings struct {
	Enabled         bool
	SMTPUsername    string
	SMTPPassword    string
	SMTPServer      string
	SMTPPort        int
	FromAddress     string
	RecipientAdress string
}

type EventbotConfig struct {
	Timezone      string
	DBFile        string
	EmailSettings EmailSettings
}

type BotEnable struct {
	Echobot_enable  bool
	Pingbot_enable  bool
	Eventbot_enable bool
}

type Config struct {
	Bots        []BotConnection
	General     GeneralConfig
	Functions   BotEnable
	EventbotCfg EventbotConfig
}

func MkExampleConfig() Config {
	return Config{
		General: GeneralConfig{
			OwnerNick:        "myowner",
			OwnerEmailAdress: "owner@example.com",
		},
		Bots: []BotConnection{
			{
				Enabled:  true,
				Channel:  "channel0",
				Server:   "server0",
				Nick:     "nick0",
				Fullname: "fullname0",
				Quitmsg:  "quitmsg0",
			},
			{
				Enabled:  false,
				Channel:  "channel1",
				Server:   "server1",
				Nick:     "nick1",
				Fullname: "fullname1",
				Quitmsg:  "quitmsg1",
			},
		},
		Functions: BotEnable{
			Echobot_enable:  false,
			Pingbot_enable:  true,
			Eventbot_enable: true,
		},
		EventbotCfg: EventbotConfig{
			Timezone: "Europe/Berlin",
			DBFile:   "/tmp/her0ld-events.db",
		},
	}
}

func LoadConfig(filename string) (Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(filename, &cfg)
	return cfg, err
}

func SaveConfig(filename string, cfg Config) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := toml.NewEncoder(f)
	err = encoder.Encode(cfg)
	return err
}
