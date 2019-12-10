package errors

// EventError type
type EventError string

func (ee EventError) Error() string {
	return string(ee)
}

var (
	ErrOverlaping        = EventError("another event exists for this date")
	ErrEventExist        = EventError("event alredy exist")
	ErrEventNotFound     = EventError("not found event")
	ErrWrangParams       = EventError("invalid input params")
	ErrConfigWrangParams = EventError("can not validate config file")
)
