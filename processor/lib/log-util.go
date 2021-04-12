package lib

import (
	"backend/gen/proto/base"
	log "github.com/sirupsen/logrus"
)

func PageRefLogger(pageRef *base.PageRef, operation string) *log.Entry {
	return log.WithFields(map[string]interface{}{
		"pageRefId":   pageRef.Id,
		"url":         pageRef.Url,
		"state":       pageRef.State.String(),
		"status":      pageRef.Status.String(),
		"websiteName": pageRef.WebsiteName,
		"operation":   operation,
	})
}
