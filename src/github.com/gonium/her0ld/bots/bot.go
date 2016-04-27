package her0ld

// todo: create a type for bot responses (instead of []string)

/* this is the interface that all bots must comply to. */
type Bot interface {
	ProcessChannelEvent(source, sender, message string) []string
}
