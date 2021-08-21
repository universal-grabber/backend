package common

import (
	"backend/api/model"
	"backend/gen/proto/base"
)

func PageRefRecordToTags(pageRef model.PageRef) map[string]string {
	var tags = make(map[string]string)

	tags["state"] = pageRef.Data.State
	tags["status"] = pageRef.Data.Status
	tags["source"] = pageRef.Data.Source

	return tags
}


func PageRefRecordToTags2(pageRef base.PageRef) map[string]string {
	var tags = make(map[string]string)

	tags["state"] = pageRef.State.String()
	tags["status"] = pageRef.Status.String()
	tags["source"] = pageRef.WebsiteName

	return tags
}
