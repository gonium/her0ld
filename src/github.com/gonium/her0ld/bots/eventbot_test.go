package her0ldbot

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func MkEventBot() EventBot {
	file, _ := ioutil.TempFile(os.TempDir(), "her0ld-eventbot-test")
	return NewEventBot("EventBot", file, "Europe/Berlin")
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
