package alerts

// Sender defines an interface that implements a generic alert sender.
type Sender interface {
	Send(payload []byte, memo string, recipients []string) error
}
