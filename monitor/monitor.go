package monitor

// Monitor defines an interface that is responsible for monitoring for a
// specific event that will ultimately trigger a potential alert.
//
// TODO: Introduce flexible filters
type Monitor interface {
	Name() string
	Memo() string
	Exec() (res, id []byte, err error)
}
