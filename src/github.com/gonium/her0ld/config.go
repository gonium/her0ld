package her0ld

import (
	"github.com/BurntSushi/toml"
	"os"
)

type BotConnection struct {
	Enabled  bool
	Channel  string
	Server   string
	Nick     string
	Fullname string
	Quitmsg  string
}

type BotEnable struct {
	Echobot_enable bool
	Pingbot_enable bool
}

type Config struct {
	Bots      []BotConnection
	Functions BotEnable
}

func MkExampleConfig() Config {
	return Config{
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
			Echobot_enable: false,
			Pingbot_enable: false,
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
