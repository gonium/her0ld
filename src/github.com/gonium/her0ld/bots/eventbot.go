package her0ldbot

import (
	"bytes"
	"fmt"
	"github.com/gonium/her0ld"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/smtp"
	"strconv"
	"strings"
	"text/template"
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
	* upcoming - list all upcoming events (with id)
	* del <id> - remove the event with the given id
	* today - show all events today`
	EVENTBOT_CMD_ADD                  = "add"
	EVENTBOT_CMD_ADD_SUCCESS          = "Recorded new event."
	EVENTBOT_CMD_LIST                 = "upcoming"
	EVENTBOT_CMD_LIST_NONE_AVAILABLE  = "no upcoming events."
	EVENTBOT_CMD_TODAY                = "today"
	EVENTBOT_CMD_TODAY_NONE_AVAILABLE = "no events today."
	EVENTBOT_CMD_DELETE               = "del"
	EVENTBOT_CMD_EVENT_UNKNOWN        = "Unknown event."
	EVENTBOT_CMD_DELETED_EVENT        = "Event %d deleted."
	EVENTBOT_CMD_MAILTEST             = "mailtest"
	EVENTBOT_MAILTEST_REPLY           = "Attempted to send test mail."
)

/********************************** Event *************************************/
type Event struct {
	Id          int `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Starttime   time.Time
	Description string
}

func (e *Event) String() string {
	return fmt.Sprintf("%d - %s: %s", e.Id,
		e.Starttime.Format("Mon 2 Jan 2006, 15:04:05"),
		e.Description)
}

/********************************** MailSender ***********************************/
type MailSender struct {
	FromAddress string
	ToAddress   string
	SMTPAuth    smtp.Auth
	SMTPServer  string
	SMTPPort    int
}

func NewMailSender(from, to, username, password, server string, port int) *MailSender {
	return &MailSender{
		FromAddress: from,
		ToAddress:   to,
		SMTPAuth:    smtp.PlainAuth("", username, password, server),
		SMTPServer:  server,
		SMTPPort:    port,
	}
}

func (ms *MailSender) SendPlainTextMail(msg string) {
	err := smtp.SendMail(ms.SMTPServer+":"+strconv.Itoa(ms.SMTPPort),
		ms.SMTPAuth,
		ms.FromAddress,
		[]string{ms.ToAddress},
		[]byte(msg))
	if err != nil {
		log.Print("ERROR: attempting to send a mail ", err)
	}
}

func (ms *MailSender) SendEventList(events []Event) {
	type SmtpTemplateData struct {
		From         string
		To           string
		Subject      string
		PrimaryEvent string
		OtherEvents  string
	}
	const emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}

Hallo,

kurze Erinnerung: Heute findet

{{.PrimaryEvent}}

statt. Weitere geplante Events:

{{.OtherEvents}}

Lieben Gruß,
{{.From}}
`
	var doc bytes.Buffer
	var err error
	// TODO: Load Template from configuration file
	context := &SmtpTemplateData{ms.FromAddress,
		ms.ToAddress,
		"Heutige Events beim Chaos inKL.",
		"Primärevent",
		"Andere Events"}
	t := template.New("emailTemplate")
	if t, err = t.Parse(emailTemplate); err != nil {
		log.Print("error trying to parse mail template ", err)
	}
	if err = t.Execute(&doc, context); err != nil {
		log.Print("error trying to execute mail template ", err)
	}
	ms.SendPlainTextMail(doc.String())
}

/********************************** EventBot *************************************/
/* The EventBot maintains a list of events and attempts to
 * remind people. */
type EventBot struct {
	BotName            string
	TimeLocation       *time.Location
	NumMessagesHandled int
	Db                 *gorm.DB
	MailSender         *MailSender
}

func NewEventBot(name string, cfg her0ld.EventbotConfig) *EventBot {
	if cfg.DBFile == "" {
		log.Fatalf("Eventbot: No database file given - aborting")
	}
	db, err := gorm.Open("sqlite3", cfg.DBFile)
	// no sense in continuing w/o database
	if err != nil {
		log.Fatalf("Eventbot: Cannot open database - %s", err.Error())
	}
	db.AutoMigrate(&Event{})
	if cfg.Timezone == "" {
		log.Fatalf("Eventbot: No timezone configured - aborting.")
	}
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("Eventbot: Cannot load location based on timezone - %s", err.Error())
	}
	ms := NewMailSender(
		cfg.EmailSettings.FromAddress,
		cfg.EmailSettings.ToAddress,
		cfg.EmailSettings.SMTPUsername,
		cfg.EmailSettings.SMTPPassword,
		cfg.EmailSettings.SMTPServer,
		cfg.EmailSettings.SMTPPort,
	)
	return &EventBot{
		BotName:            name,
		TimeLocation:       loc,
		NumMessagesHandled: 0,
		Db:                 db,
		MailSender:         ms,
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
		var answer []string
		var events []Event
		b.Db.Where("starttime > ?", time.Now()).Find(&events)
		if len(events) == 0 {
			answer = append(answer, EVENTBOT_CMD_LIST_NONE_AVAILABLE)
		} else {
			for _, event := range events {
				answer = append(answer, event.String())
			}
		}
		return b.strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_TODAY)) {
		var answer []string
		var events []Event
		starttime := time.Now().Truncate(24 * time.Hour)
		endtime := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
		b.Db.Where("starttime > ? and starttime < ?", starttime, endtime).Find(&events)
		if len(events) == 0 {
			answer = append(answer, EVENTBOT_CMD_TODAY_NONE_AVAILABLE)
		} else {
			for _, event := range events {
				answer = append(answer, event.String())
			}
		}
		return b.strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_DELETE)) {
		request := strings.Split(msg.Message, " ")
		var answer []string
		var events []Event
		b.Db.Where("ID == ?", request[2]).Find(&events)
		if len(events) == 0 {
			answer = append(answer, EVENTBOT_CMD_EVENT_UNKNOWN)
		} else {
			for _, event := range events {
				b.Db.Delete(&event)
				answer = append(answer, fmt.Sprintf(EVENTBOT_CMD_DELETED_EVENT,
					event.Id))
			}
		}
		return b.strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_MAILTEST)) {
		mailtext := "To: md@gonium.net\r\n" +
			"Subject: Plain test Mail\r\n" +
			"\r\n" +
			"Testing mail.\r\n"
		b.MailSender.SendPlainTextMail(mailtext)
		answer := []string{EVENTBOT_MAILTEST_REPLY}
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

// TODO:
// - Generate Email templates and process them.
// (https://gist.github.com/nathanleclaire/8662755)
// - Send reminder emails based on cron
// (https://godoc.org/github.com/robfig/cron).
