package helper

import (
	"backend/api/model"
	log "github.com/sirupsen/logrus"
)

func PageRefLogger(pageRef *model.PageRef, operation string) *log.Entry {
	return log.WithFields(map[string]interface{}{
		"pageRefId":   pageRef.Id.String(),
		"url":         pageRef.Data.Url,
		"state":       pageRef.Data.State,
		"status":      pageRef.Data.Status,
		"websiteName": pageRef.Data.Source,
		"operation":   operation,
	})
}
