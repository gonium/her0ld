package her0ldbot

import (
	"testing"
)

func TestIgnoreNoCommandMsg(t *testing.T) {
	bot := NewPingBot("Pingbot")
	initialcounter := bot.NumMessagesHandled
	texts := [4]string{"djewoijdowe", "foobar", "!pingoooo", "ping"}
	for _, text := range texts {
		msg := InboundMessage{
			Channel: "test",
			Nick:    "testnick",
			Message: text,
		}
		answerlines, err := bot.ProcessChannelEvent(msg)
		if err != nil {
			t.Fatalf("Bot should process noncommand messages without error, msg=%s", text)
		}
		if answerlines != nil {
			t.Fatalf("Bot should ignore lines not equal to !ping, msg=%s", text)
		}
	}
	if bot.NumMessagesHandled != initialcounter {
		t.Fatalf("The bot handled %d messages - expected none.",
			bot.NumMessagesHandled)
	}
}

func TestValidCommandMsg(t *testing.T) {
	bot := NewPingBot("Pingbot")
	texts := [2]string{"!ping", "!PING"}
	for _, text := range texts {
		msg := InboundMessage{
			Channel: "test",
			Nick:    "testnick",
			Message: text,
		}
		answerlines, err := bot.ProcessChannelEvent(msg)
		if err != nil {
			t.Fatalf("Bot should process command messages without error, msg=%s", text)
		}
		if answerlines == nil {
			t.Fatalf("Bot should respond with PONG message, msg=%s",
				text)
		}
		if len(answerlines) != 1 {
			t.Fatalf("Bot should respond with exactly one answer line")
		}
		if answerlines[0].Message != "PONG" {
			t.Fatalf("Bot did respond with %s, expected PONG",
				answerlines[0].Message)
		}
	}
	if bot.NumMessagesHandled != len(texts) {
		t.Fatalf("The bot handled %d messages - expected %d.",
			bot.NumMessagesHandled, len(texts))
	}
}
