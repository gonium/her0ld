package her0ldbot

import (
	"fmt"
	"github.com/gonium/her0ld"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var genConfig her0ld.GeneralConfig = her0ld.GeneralConfig{
	OwnerNick:        "testnick",
	OwnerEmailAdress: "testowner@example.com",
}

func MkEventBot() *EventBot {
	file := filepath.Join(os.TempDir(), "her0ld-eventbot-test.db")
	// delete test db if the file exists
	_ = os.Remove(file)
	return NewEventBot(
		"EventBot",
		her0ld.EventbotConfig{
			DBFile:   file,
			Timezone: "Europe/Berlin",
		},
		genConfig,
	)
}

// Ignore ordinary chat messages
func TestNoCommand(t *testing.T) {
	bot := MkEventBot()
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: "ordinarychatline",
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("Invalid command triggered unexpected error: %s",
			err.Error())
	} else {
		if len(response) != 0 {
			t.Fatalf("Invalid response length - expected %d, got %d", 0,
				len(response))
		}
	}
}

// Ignore an invalid command (i.e. for another bot)
func TestInvalidCommand(t *testing.T) {
	bot := MkEventBot()
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: "!invalidcmd",
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("Invalid command triggered unexpected error: %s",
			err.Error())
	} else {
		if len(response) != 0 {
			t.Fatalf("Invalid response length - expected %d, got %d", 0,
				len(response))
		}
	}
}

func TestEmptyCommand(t *testing.T) {
	bot := MkEventBot()
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: EVENTBOT_PREFIX,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("Invalid command triggered unexpected error: %s",
			err.Error())
	} else {
		if len(response) != 1 {
			t.Fatalf("Invalid response length - expected %d, got %d", 1,
				len(response))
		} else {
			if response[0].Destination != msg.Channel {
				t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
					response[0].Destination)
			}
			if response[0].Message != EVENTBOT_INVALID_COMMAND {
				t.Fatalf("Invalid response: Expected >%s<, got >%s<",
					EVENTBOT_INVALID_COMMAND, response[0].Message)
			}
		}
	}
}

func TestHelpCommand(t *testing.T) {
	bot := MkEventBot()
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX, EVENTBOT_CMD_HELP)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("Invalid command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_HELP_TEXT, "\n")
		expected_len := len(expected)
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

// Test mail requested by owner of the bot
func TestValidOwnerMail(t *testing.T) {
	bot := MkEventBot()
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX, EVENTBOT_CMD_MAILTEST)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    genConfig.OwnerNick,
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("mailtest command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_MAILTEST_REPLY, "\n")
		expected_len := len(expected)
		if len(response) != expected_len {
			t.Fatalf("Invalid response length - expected %d, got %d", expected_len,
				len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

// Test mail requested by not-an-owner of the bot
func TestInvalidOwnerMail(t *testing.T) {
	bot := MkEventBot()
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX, EVENTBOT_CMD_MAILTEST)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    genConfig.OwnerNick + "fooo",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("mailtest command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_MAILTEST_NOTAUTHORIZED, "\n")
		expected_len := len(expected)
		if len(response) != expected_len {
			t.Fatalf("Invalid response length - expected %d, got %d", expected_len,
				len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestInvalidAddCommand(t *testing.T) {
	bot := MkEventBot()
	line := fmt.Sprintf("%s %s invalid date event", EVENTBOT_PREFIX, EVENTBOT_CMD_ADD)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("Invalid command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_INVALID_TIME_FORMAT, "\n")
		expected_len := 2 // two lines
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestValidAddCommand(t *testing.T) {
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	line := fmt.Sprintf("%s %s %s test", EVENTBOT_PREFIX,
		EVENTBOT_CMD_ADD, EVENTBOT_TIME_FORMAT)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_CMD_ADD_SUCCESS, "\n")
		expected_len := 1
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
	// Check wether we have an additional event in the database
	bot.Db.Find(&events)
	if len(events) != 1 {
		t.Fatalf("Event not stored in db.")
	}
}

func TestListNoAvailableEvents(t *testing.T) {
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX,
		EVENTBOT_CMD_LIST)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_CMD_LIST_NONE_AVAILABLE, "\n")
		expected_len := 1
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestListOnlyUpcomingEvents(t *testing.T) {
	// Create two events - one in the past, one in the future. list should
	// only display the future one.
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	oldEvent := Event{
		Starttime:   time.Now().Add(-10 * time.Second),
		Description: "Old legacy event",
	}
	upcomingEvent := Event{
		Starttime:   time.Now().Add(10 * time.Second),
		Description: "Upcoming legacy event",
	}
	bot.Db.Create(&oldEvent)
	bot.Db.Create(&upcomingEvent)
	bot.Db.Find(&events)
	if len(events) != 2 {
		t.Fatalf("Failed to create test events.")
	}
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX,
		EVENTBOT_CMD_LIST)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(upcomingEvent.String(), "\n")
		expected_len := 1
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestListTodayNoEvents(t *testing.T) {
	// Create two events - one in the past, one in the future. list should
	// only display the future one.
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	oldEvent := Event{
		Starttime:   time.Now().Add(-10 * 24 * time.Hour),
		Description: "Old legacy event",
	}
	upcomingEvent := Event{
		Starttime:   time.Now().Add(10 * 24 * time.Hour),
		Description: "Upcoming legacy event",
	}
	bot.Db.Create(&oldEvent)
	bot.Db.Create(&upcomingEvent)
	bot.Db.Find(&events)
	if len(events) != 2 {
		t.Fatalf("Failed to create test events.")
	}
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX,
		EVENTBOT_CMD_TODAY)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_CMD_TODAY_NONE_AVAILABLE, "\n")
		expected_len := 1
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestListTodayTwoEvents(t *testing.T) {
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	oldEvent := Event{
		Starttime:   time.Now().Add(-10 * 24 * time.Hour),
		Description: "Old legacy event",
	}
	firstEvent := Event{
		Starttime:   time.Now(),
		Description: "first today event",
	}
	secondEvent := Event{
		Starttime:   time.Now().Add(10 * time.Second),
		Description: "second today event",
	}
	upcomingEvent := Event{
		Starttime:   time.Now().Add(10 * 24 * time.Hour),
		Description: "Upcoming legacy event",
	}
	bot.Db.Create(&oldEvent)
	bot.Db.Create(&firstEvent)
	bot.Db.Create(&secondEvent)
	bot.Db.Create(&upcomingEvent)
	bot.Db.Find(&events)
	if len(events) != 4 {
		t.Fatalf("Failed to create test events.")
	}
	line := fmt.Sprintf("%s %s", EVENTBOT_PREFIX,
		EVENTBOT_CMD_TODAY)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		eventlist := fmt.Sprintf("%s\n%s", firstEvent.String(),
			secondEvent.String())
		expected := strings.Split(eventlist, "\n")
		expected_len := len(expected)
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestDeleteInvalidEventID(t *testing.T) {
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	firstEvent := Event{
		Starttime:   time.Now(),
		Description: "first today event",
	}
	secondEvent := Event{
		Starttime:   time.Now().Add(10 * time.Second),
		Description: "second today event",
	}
	bot.Db.Create(&firstEvent)
	bot.Db.Create(&secondEvent)
	bot.Db.Find(&events)
	if len(events) != 2 {
		t.Fatalf("Failed to create test events.")
	}
	line := fmt.Sprintf("%s %s 23", EVENTBOT_PREFIX,
		EVENTBOT_CMD_DELETE)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(EVENTBOT_CMD_EVENT_UNKNOWN, "\n")
		expected_len := len(expected)
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
}

func TestDeleteEventID(t *testing.T) {
	bot := MkEventBot()
	var events []Event
	bot.Db.Find(&events)
	if len(events) != 0 {
		t.Fatalf("Db not empty.")
	}
	firstEvent := Event{
		Starttime:   time.Now(),
		Description: "first today event",
	}
	secondEvent := Event{
		Starttime:   time.Now().Add(10 * time.Second),
		Description: "second today event",
	}
	bot.Db.Create(&firstEvent)
	bot.Db.Create(&secondEvent)
	bot.Db.Find(&events)
	if len(events) != 2 {
		t.Fatalf("Failed to create test events.")
	}
	line := fmt.Sprintf("%s %s 1", EVENTBOT_PREFIX,
		EVENTBOT_CMD_DELETE)
	msg := InboundMessage{
		Channel: "Channel",
		Nick:    "Nick",
		Message: line,
	}
	response, err := bot.ProcessChannelEvent(msg)
	if err != nil {
		t.Fatalf("command triggered unexpected error: %s",
			err.Error())
	} else {
		expected := strings.Split(fmt.Sprintf(EVENTBOT_CMD_DELETED_EVENT, 1), "\n")
		expected_len := len(expected)
		if len(response) != expected_len {
			t.Fatalf("Invalid length of response %#v - expected %d, got %d",
				response, expected_len, len(response))
		} else {
			for idx, expected_line := range expected {
				if response[idx].Destination != msg.Channel {
					t.Fatalf("Invalid destination: Expected >%s<, got >%s<", msg.Channel,
						response[idx].Destination)
				}
				if response[idx].Message != expected_line {
					t.Fatalf("Invalid response: Expected >%s<, got >%s<",
						expected_line, response[idx].Message)
				}
			}
		}
	}
	bot.Db.Find(&events)
	if len(events) != 1 {
		t.Fatalf("Failed to delete test event.")
	}

}
