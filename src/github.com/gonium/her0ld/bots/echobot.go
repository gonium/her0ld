package her0ldbot

import (
	"fmt"
)

/* The EchoBot simply echoes user inputs. It is not really
 * useful except for testing and development. */
type EchoBot struct {
	BotName            string
	NumMessagesHandled int
}

func NewEchoBot(name string) *EchoBot {
	return &EchoBot{BotName: name, NumMessagesHandled: 0}
}

func (b *EchoBot) ProcessChannelEvent(msg InboundMessage) ([]OutboundMessage, error) {
	b.NumMessagesHandled += 1
	answer := fmt.Sprintf("%s: %s", msg.Nick, msg.Message)
	reply := make([]OutboundMessage, 1)
	reply[0] = OutboundMessage{Destination: msg.Channel, Message: answer}
	return reply, nil
}

func (b *EchoBot) ProcessQueryEvent(msg InboundMessage) ([]OutboundMessage, error) {
	b.NumMessagesHandled += 1
	answer := fmt.Sprintf("%s", msg.Message)
	reply := make([]OutboundMessage, 1)
	reply[0] = OutboundMessage{Destination: msg.Channel, Message: answer}
	return reply, nil
}

func (b *EchoBot) GetName() string {
	return b.BotName
}
