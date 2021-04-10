package model

type Exception struct {
	message string
}

func (e *Exception) Error() string {
	return e.message
}

func NewException(message string) *Exception {
	exception := new(Exception)

	exception.message = message

	return exception
}
