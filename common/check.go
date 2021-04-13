package common

import (
	"backend/gen/proto/base"
	log "github.com/sirupsen/logrus"
)

func Check(err error) {
	if err != nil {
		log.WithField("operation", "fatal-error").
			Panic(err)
	}
}

func CheckWithPageRef(err error, pageRef *base.PageRef) {
	if err != nil {
		PageRefLogger(pageRef, "fatal-error").
			WithError(err).
			Errorf("error: %s", err)

	}
}

func CheckWithPageRefLogOnly(err error, pageRef *base.PageRef) {
	if err != nil {
		PageRefLogger(pageRef, "fatal-error").
			WithError(err).
			Errorf("error: %s", err)
	}
}
