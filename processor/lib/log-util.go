package lib

import (
	"backend/processor/model"
	log "github.com/sirupsen/logrus"
)

func PageRefLogger(pageRef *model.PageRef, operation string) *log.Entry {
	return log.WithFields(map[string]interface{}{
		"pageRefId":   pageRef.Id.String(),
		"url":         pageRef.Url,
		"state":       pageRef.State,
		"status":      pageRef.Status,
		"websiteName": pageRef.WebsiteName,
		"operation":   operation,
	})
}
