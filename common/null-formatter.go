package common

import log "github.com/sirupsen/logrus"

type NullFormatter struct {
}

// Don't spend time formatting logs
func (NullFormatter) Format(e *log.Entry) ([]byte, error) {
	return []byte{}, nil
}
