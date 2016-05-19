package her0ldbot

import (
	"fmt"
	"strings"
)

const (
	EVENTBOT_INVALID_COMMAND = "invalid command - see !event help."
	EVENTBOT_PREFIX          = "!event"
	EVENTBOT_CMD_HELP        = "help"
	EVENTBOT_HELP_TEXT       = `!event usage:
	* help - this text
	* add <date> <description> - add an event
	* list - list all events (with id)
	* del <id> - remove the event with the given id
	* today - show all events today`
)

/* The EventBot maintains a list of events and attempts to
 * remind people. */
type EventBot struct {
	BotName            string
	NumMessagesHandled int
}

func NewEventBot(name string) *EventBot {
	return &EventBot{BotName: name, NumMessagesHandled: 0}
}

func (b *EventBot) strings2reply(dest string, lines []string) []OutboundMessage {
	reply := make([]OutboundMessage, len(lines))
	for idx, line := range lines {
		reply[idx] = OutboundMessage{
			Destination: dest,
			Message:     line,
		}
	}
	return reply
}

func (b *EventBot) ProcessChannelEvent(msg InboundMessage) ([]OutboundMessage, error) {
	b.NumMessagesHandled += 1
	// !event help command
	if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_HELP)) {
		answer := strings.Split(EVENTBOT_HELP_TEXT, "\n")
		return b.strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, EVENTBOT_PREFIX) { // invalid command (!event foo)
		answer := []string{EVENTBOT_INVALID_COMMAND}
		return b.strings2reply(msg.Channel, answer), nil
	} else { // something else
		return nil, nil
	}
}

func (b *EventBot) ProcessQueryEvent(msg InboundMessage) ([]OutboundMessage, error) {
	b.NumMessagesHandled += 1
	reply := make([]OutboundMessage, 1)
	reply[0] = OutboundMessage{Destination: msg.Channel, Message: "Not implemented"}
	return reply, nil
}

func (b *EventBot) GetName() string {
	return b.BotName
}
