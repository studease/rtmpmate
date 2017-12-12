package Event

import ()

// Not sure about this.
type Event struct {
	Type   byte
	Length [3]byte
	Data   []byte
}

func New() (*Event, error) {
	var e Event
	return &e, nil
}
