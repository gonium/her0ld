package her0ldbot

import (
	"fmt"
	"testing"
)

func TestValidMsg(t *testing.T) {
	bot := NewEchoBot("Echobot")
	texts := [2]string{"!ping", "foobar"}
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
			t.Fatalf("Bot should respond with an echo of the message, msg=%s",
				text)
		}
		if len(answerlines) != 1 {
			t.Fatalf("Bot should respond with exactly one answer line")
		}
		if answerlines[0].Message != fmt.Sprintf("%s: %s", msg.Nick, text) {
			t.Fatalf("Bot did respond with %s, expected PONG",
				answerlines[0].Message)
		}
	}
	if bot.NumMessagesHandled != len(texts) {
		t.Fatalf("The bot handled %d messages - expected %d.",
			bot.NumMessagesHandled, len(texts))
	}
}
