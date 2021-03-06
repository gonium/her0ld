package her0ldbot

import (
	"fmt"
	"strings"
)

/* a bot receives this type of message. */
type InboundMessage struct {
	Channel string
	Nick    string
	Message string
}

func (b InboundMessage) String() string {
	return fmt.Sprintf("C: %s: N: %s - M: %s", b.Channel, b.Nick, b.Message)
}

func (b InboundMessage) IsChannelEvent() bool {
	return strings.HasPrefix(b.Channel, "#")
}

/* a bot responds with messages of this type. */
type OutboundMessage struct {
	Destination string
	Message     string
}

/* this is the interface that all bots must comply to. */
type Bot interface {
	ProcessChannelEvent(incoming InboundMessage) ([]OutboundMessage, error)
	ProcessQueryEvent(incoming InboundMessage) ([]OutboundMessage, error)
	GetName() string
	GetHelpLines() []string
}

func strings2reply(dest string, lines []string) []OutboundMessage {
	reply := make([]OutboundMessage, len(lines))
	for idx, line := range lines {
		reply[idx] = OutboundMessage{
			Destination: dest,
			Message:     line,
		}
	}
	return reply
}
