package backend

type Event struct{ Value any }

type Backend interface {
	Size() (width, height int, err error)
	Events() <-chan Event
	Write([]byte) error
	Close() error
}
