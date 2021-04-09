package model

import uuid "github.com/satori/go.uuid"

type PageRef struct {
	Id          uuid.UUID     `bson:"_id" json:"id"`
	WebsiteName string        `bson:"websiteName" json:"websiteName"`
	Url         string        `json:"url"`
	State       string        `json:"state"`
	Status      PageRefStatus `json:"status"`
	Tags        *[]string     `json:"tags"`
}

type PageRefStatus string

const (
	PENDING   PageRefStatus = "PENDING"
	EXECUTING PageRefStatus = "EXECUTING"
	FINISHED  PageRefStatus = "FINISHED"
	FAILED    PageRefStatus = "FAILED"
)
