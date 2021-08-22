package model

import uuid "github.com/satori/go.uuid"

type PageRef struct {
	Id    uuid.UUID   `bson:"_id" json:"id"`
	Title string      `json:"title"`
	Data  PageRefData `json:"data"`
}

type PageRefData struct {
	Source  string      `bson:"websiteName" json:"websiteName"`
	Url     string      `json:"url"`
	State   string      `json:"state"`
	Status  string      `json:"status"`
	Tags    *[]string   `bson:"tags" json:"tags"`
	Options interface{} `bson:"options" json:"options"`
}
