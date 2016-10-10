package her0ldbot

import (
	"bytes"
	"fmt"
	"github.com/gonium/her0ld"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/robfig/cron"
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
	EVENTBOT_HELP_TEXT           = `* !event help - this text
* !event add <date> <description> - add an event
* !event list - list all upcoming events (with id)
* !event del <id> - remove the event with the given id
* !event today - show all events today`
	EVENTBOT_CMD_ADD                     = "add"
	EVENTBOT_CMD_ADD_SUCCESS             = "Recorded new event."
	EVENTBOT_CMD_LIST                    = "list"
	EVENTBOT_CMD_LIST_NONE_AVAILABLE     = "no upcoming events."
	EVENTBOT_CMD_TODAY                   = "today"
	EVENTBOT_CMD_TODAY_NONE_AVAILABLE    = "no events today."
	EVENTBOT_CMD_DELETE                  = "del"
	EVENTBOT_CMD_EVENT_UNKNOWN           = "Unknown event."
	EVENTBOT_CMD_DELETED_EVENT           = "Event %d deleted."
	EVENTBOT_CMD_MAILTEST                = "mailtest"
	EVENTBOT_MAILTEST_REPLY              = "Attempted to send test mail."
	EVENTBOT_MAILTEST_NOTAUTHORIZED      = "Only my owner can do this."
	EVENTBOT_CMD_MAILREMINDER            = "mailreminder"
	EVENTBOT_MAILREMINDER_REPLY          = "Attempted to send a reminder mail."
	EVENTBOT_MAILREMINDER_NOTAUTHORIZED  = "Only my owner can do this."
	EVENTBOT_MAILREMINDER_SEND_ERROR     = "Failed to send reminder email, check log."
	EVENTBOT_MAILREMINDER_NONE_AVAILABLE = "No events for today found, not sending."
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

type EventList []Event

func (e EventList) String() string {
	var eventstrings []string
	for _, s := range e {
		eventstrings = append(eventstrings, s.String())
	}
	return strings.Join(eventstrings, "\n")
}

/********************************** MailSender ***********************************/
type MailSender struct {
	FromAddress string
	SMTPAuth    smtp.Auth
	SMTPServer  string
	SMTPPort    int
}

func NewMailSender(from, username, password, server string, port int) *MailSender {
	return &MailSender{
		FromAddress: from,
		SMTPAuth:    smtp.PlainAuth("", username, password, server),
		SMTPServer:  server,
		SMTPPort:    port,
	}
}

func (ms *MailSender) SendPlainTextMail(msg string, toadress string) {
	serveradress := ms.SMTPServer + ":" + strconv.Itoa(ms.SMTPPort)
	err := smtp.SendMail(serveradress,
		ms.SMTPAuth,
		ms.FromAddress,
		[]string{toadress},
		[]byte(msg))
	if err != nil {
		log.Println("ERROR: failed to send email, ", err.Error())
	}
}

/********************************** EventBot *************************************/
/* The EventBot maintains a list of events and attempts to
 * remind people. */
type EventBot struct {
	BotName               string
	TimeLocation          *time.Location
	NumMessagesHandled    int
	Db                    *gorm.DB
	MailSender            *MailSender
	OwnerNick             string
	OwnerEmailAddress     string
	RecipientAddress      string
	EventListMailTemplate string
	Cron                  *cron.Cron
}

func NewEventBot(name string, cfg her0ld.EventbotConfig, generalcfg her0ld.GeneralConfig) *EventBot {
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
		cfg.EmailSettings.SMTPUsername,
		cfg.EmailSettings.SMTPPassword,
		cfg.EmailSettings.SMTPServer,
		cfg.EmailSettings.SMTPPort,
	)
	retval := &EventBot{
		BotName:               name,
		TimeLocation:          loc,
		NumMessagesHandled:    0,
		Db:                    db,
		MailSender:            ms,
		OwnerNick:             generalcfg.OwnerNick,
		OwnerEmailAddress:     generalcfg.OwnerEmailAddress,
		RecipientAddress:      cfg.EmailSettings.RecipientAddress,
		EventListMailTemplate: cfg.EmailSettings.EventListMailTemplate,
		Cron: cron.New(),
	}
	retval.Cron.AddFunc("0 0 1 * * *", func() {
		log.Println("Cron: Triggering event list email.")
		switch retval.SendEventList() {
		case NO_EVENT_TODAY:
			log.Printf("Cron: %s", EVENTBOT_MAILREMINDER_NONE_AVAILABLE)
		case SEND_ERROR:
			log.Printf("Cron: %s", EVENTBOT_MAILREMINDER_SEND_ERROR)
		case SEND_SUCCESS:
			log.Printf("Cron: %s", EVENTBOT_MAILREMINDER_REPLY)
		default:
			log.Println("Cron: Unknown return code from SendEventList - WTF?")
		}
	})
	retval.Cron.Start()
	return retval
}

type SendEventListStatus int

const (
	NO_EVENT_TODAY = iota
	SEND_SUCCESS   = iota
	SEND_ERROR     = iota
)

func (eb *EventBot) SendEventList() (status SendEventListStatus) {
	var todayEvents EventList
	todaytime := time.Now().Truncate(24 * time.Hour)
	tomorrow := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	eb.Db.Where("starttime > ? and starttime < ?", todaytime, tomorrow).Find(&todayEvents)
	if len(todayEvents) == 0 {
		status = NO_EVENT_TODAY
	} else {
		var upcomingEvents EventList
		eb.Db.Where("starttime > ?", tomorrow).Find(&upcomingEvents)
		type TemplateData struct {
			From            string
			To              string
			Now             string
			Subject         string
			HighlightEvents string
			UpcomingEvents  string
		}
		emailTemplate := `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
Date: {{.Now}}
Content-Type: text/plain; charset=UTF-8

` + eb.EventListMailTemplate
		var doc bytes.Buffer
		var err error
		// TODO: Load Template from configuration file
		context := &TemplateData{
			From:            eb.MailSender.FromAddress,
			To:              eb.RecipientAddress,
			Now:             time.Now().Format(time.RFC822),
			Subject:         "Reminder: Heute beim Chaos inKL.",
			HighlightEvents: todayEvents.String(),
			UpcomingEvents:  upcomingEvents.String(),
		}
		t := template.New("emailTemplate")
		if t, err = t.Parse(emailTemplate); err != nil {
			log.Print("error trying to parse mail template ", err)
			status = SEND_ERROR
		}
		if err = t.Execute(&doc, context); err != nil {
			log.Print("error trying to execute mail template ", err)
			status = SEND_ERROR
		} else {
			eb.MailSender.SendPlainTextMail(doc.String(), eb.RecipientAddress)
			status = SEND_SUCCESS
		}
	}
	return status
}

func (b *EventBot) isFromOwner(msg InboundMessage) bool {
	return msg.Nick == b.OwnerNick
}

func (b *EventBot) ProcessChannelEvent(msg InboundMessage) ([]OutboundMessage, error) {
	b.NumMessagesHandled += 1
	// look for commands
	//log.Printf("Processing message %s", msg.Message)
	if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_HELP)) {
		// help command
		answer := strings.Split(EVENTBOT_HELP_TEXT, "\n")
		return strings2reply(msg.Channel, answer), nil
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
			return strings2reply(msg.Channel, answer), nil
		} else {
			description := strings.Join(request[3:], " ")
			newEvent := Event{
				Starttime:   date,
				Description: description,
			}
			b.Db.Create(&newEvent)
			answer := []string{EVENTBOT_CMD_ADD_SUCCESS}
			return strings2reply(msg.Channel, answer), nil
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
		return strings2reply(msg.Channel, answer), nil
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
		return strings2reply(msg.Channel, answer), nil
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
		return strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_MAILTEST)) {
		answer := []string{EVENTBOT_MAILTEST_NOTAUTHORIZED}
		if b.isFromOwner(msg) {
			mailtext := "To: " + b.OwnerEmailAddress + "\r\n" +
				"From: " + b.MailSender.FromAddress + "\r\n" +
				"Date: " + (time.Now().Format(time.RFC822)) + "\r\n" +
				"Subject: her0ld mail test\r\n" +
				"\r\n" +
				"Testing mail. If you can read this everything should be working.\r\n"
			go b.MailSender.SendPlainTextMail(mailtext, b.OwnerEmailAddress)
			answer = []string{EVENTBOT_MAILTEST_REPLY}
		}
		return strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, fmt.Sprintf("%s %s",
		EVENTBOT_PREFIX, EVENTBOT_CMD_MAILREMINDER)) {
		answer := []string{EVENTBOT_MAILREMINDER_NOTAUTHORIZED}
		if b.isFromOwner(msg) {
			switch b.SendEventList() {
			case NO_EVENT_TODAY:
				answer = []string{EVENTBOT_MAILREMINDER_NONE_AVAILABLE}
			case SEND_ERROR:
				answer = []string{EVENTBOT_MAILREMINDER_SEND_ERROR}
			case SEND_SUCCESS:
				answer = []string{EVENTBOT_MAILREMINDER_REPLY}
			default:
				log.Println("Unknown return code from SendEventList - WTF?")
			}
		}
		return strings2reply(msg.Channel, answer), nil
	} else if strings.HasPrefix(msg.Message, EVENTBOT_PREFIX) {
		// invalid command (!event foo)
		answer := []string{EVENTBOT_INVALID_COMMAND}
		return strings2reply(msg.Channel, answer), nil
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

func (b *EventBot) GetHelpLines() []string {
	return strings.Split(EVENTBOT_HELP_TEXT, "\n")
}
