package her0ldbot

import (
	"fmt"
	"strings"
)

/* The Pingbot replies with "PONG" when it receives a "!ping" message.
 * */
type PingBot struct {
	BotName            string
	NumMessagesHandled int
}

func NewPingBot(name string) *PingBot {
	return &PingBot{BotName: name, NumMessagesHandled: 0}
}

func (b *PingBot) ProcessChannelEvent(msg InboundMessage) ([]OutboundMessage, error) {
	if strings.ToLower(msg.Message) == "!ping" {
		b.NumMessagesHandled += 1
		answer := fmt.Sprintf("PONG")
		reply := make([]OutboundMessage, 1)
		reply[0] = OutboundMessage{Destination: msg.Channel, Message: answer}
		return reply, nil
	} else {
		return nil, nil
	}
}

func (b *PingBot) ProcessQueryEvent(msg InboundMessage) ([]OutboundMessage, error) {
	return nil, fmt.Errorf("PingBot does not implement query event handling")
}

func (b *PingBot) GetName() string {
	return b.BotName
}
