package model

import uuid "github.com/satori/go.uuid"

type PageData struct {
	Id     uuid.UUID `bson:"_id" json:"id"`
	Record *Record   `bson:"record" json:"record"`
}
