package model

type PageData struct {
	Id     string  `bson:"_id" json:"id"`
	Record *Record `bson:"record" json:"record"`
}
