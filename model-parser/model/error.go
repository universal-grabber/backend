package model

type Error struct {
	Message string
	ErrorType string
}

func (e Error) Error() string {
	return e.Message
}
