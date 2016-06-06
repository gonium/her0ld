package her0ldbot

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"strings"
	"time"
)

const (
	EVENTBOT_INVALID_TIME_FORMAT = "Invalid time format, use e.g.  01.01.2016-16:00"
	EVENTBOT_TIME_FORMAT         = "02.01.2006-15:04"
	EVENTBOT_INVALID_COMMAND     = "invalid command - see !event help."
	EVENTBOT_PREFIX              = "!event"
	EVENTBOT_CMD_HELP            = "help"
	EVENTBOT_HELP_TEXT           = `!event usage:
	* help - this text
	* add <date> <description> - add an event
	* list - list all events (with id)
	* del <id> - remove the event with the given id
	* today - show all events today`
	EVENTBOT_CMD_ADD         = "add"
	EVENTBOT_CMD_ADD_SUCCESS = "Recorded new event."
	EVENTBOT_CMD_LIST        = "list"
)

/********************************** Event *************************************/
type Event struct {
	Id          int `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Starttime   time.Time
	Description string
}

func (e *Event) String() string {
	return fmt.Sprintf("%d - %s: %s", e.Id,
		e.Starttime.Format("Mon 2 Jan 15:04:05 2006"),
		e.Description)
}

/********************************** EventBot *************************************/
/* The EventBot maintains a list of events and attempts to
 * remind people. */
type EventBot struct {
	BotName            string
	TimeLocation       *time.Location
	NumMessagesHandled int
	Db                 *gorm.DB
}

func NewEventBot(name string, dbfile string, timelocation string) *EventBot {
	db, err := gorm.Open("sqlite3", dbfile)
	// no sense in continuing w/o database
	if err != nil {
		log.Fatalf("Cannot open database - %s", err.Error())
	}
	db.AutoMigrate(&Event{})
	loc, err := time.LoadLocation(timelocation)
	if err != nil {
		log.Fatalf("Cannot load location - %s", err.Error())
	}
	return &EventBot{
		BotName:            name,
		TimeLocation:       loc,
		NumMessagesHandled: 0,
		Db:                 db,
	}
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
	// look for commands
	if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_HELP)) {
		// help command
		answer := strings.Split(EVENTBOT_HELP_TEXT, "\n")
		return b.strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_ADD)) {
		// add new event command
		request := strings.Split(msg.Message, " ")
		// request[0|1] contain the add command - we need date and
		// description.
		if date, err := time.ParseInLocation(EVENTBOT_TIME_FORMAT,
			request[2], b.TimeLocation); err != nil {
			answer := []string{
				EVENTBOT_INVALID_TIME_FORMAT,
				fmt.Sprintf("Error was: %s", err.Error()),
			}
			return b.strings2reply(msg.Channel, answer), nil
		} else {
			description := strings.Join(request[3:], " ")
			newEvent := Event{
				Starttime:   date,
				Description: description,
			}
			b.Db.Create(&newEvent)
			answer := []string{EVENTBOT_CMD_ADD_SUCCESS}
			return b.strings2reply(msg.Channel, answer), nil
		}
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_LIST)) {
		var events []Event
		b.Db.Find(&events)
		var answer []string
		for _, event := range events {
			answer = append(answer, event.String())
		}
		return b.strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, EVENTBOT_PREFIX) {
		// invalid command (!event foo)
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
