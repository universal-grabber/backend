package model

import uuid "github.com/satori/go.uuid"

type PageRef struct {
	Id          uuid.UUID `bson:"_id" json:"id"`
	WebsiteName string    `bson:"websiteName" json:"websiteName"`
	Url         string    `json:"url"`
	State       string    `json:"state"`
	Status      string    `json:"status"`
	Tags        *[]string `bson:"tags" json:"tags"`
}
