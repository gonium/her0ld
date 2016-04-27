package her0ld

type EchoBot struct {
	BotName            string
	NumMessagesHandled int
}

func NewEchoBot(name string) *EchoBot {
	return &EchoBot{BotName: name, NumMessagesHandled: 0}
}

func (b *EchoBot) ProcessChannelEvent(source, sender, message string) []string {
	b.NumMessagesHandled += 1
	return nil
}
