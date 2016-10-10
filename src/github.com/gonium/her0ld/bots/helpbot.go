package her0ldbot

import (
	"fmt"
	"strings"
)

type HelpBot struct {
	BotName            string
	NumMessagesHandled int
	bots               []Bot
}

func NewHelpBot(name string, bots []Bot) *HelpBot {
	return &HelpBot{
		BotName:            name,
		NumMessagesHandled: 0,
		bots:               bots,
	}
}

func (b *HelpBot) ProcessChannelEvent(msg InboundMessage) ([]OutboundMessage, error) {
	if strings.ToLower(msg.Message) == "!help" {
		b.NumMessagesHandled += 1
		var answer []string
		for _, b := range b.bots {
			answer = append(answer, fmt.Sprintf("%s commands:", b.GetName()))
			lines := b.GetHelpLines()
			for _, l := range lines {
				answer = append(answer, l)
			}
		}
		answer = append(answer, "Find my code at https://github.com/gonium/her0ld")
		reply := make([]OutboundMessage, len(answer))
		for idx, line := range answer {
			reply[idx] = OutboundMessage{
				Destination: msg.Channel,
				Message:     line,
				//Message:     answer[idx],
			}
		}
		return reply, nil
	} else {
		return nil, nil
	}
}

func (b *HelpBot) ProcessQueryEvent(msg InboundMessage) ([]OutboundMessage, error) {
	return nil, fmt.Errorf("HelpBot does not implement query event handling")
}

func (b *HelpBot) GetName() string {
	return b.BotName
}

func (b *HelpBot) GetHelpLines() []string {
	return []string{"!help prints the help texts of all available bots."}
}
