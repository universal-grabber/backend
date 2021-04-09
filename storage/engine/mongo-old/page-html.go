package mongo_old

import uuid "github.com/satori/go.uuid"

type SimplePageHtml struct {
	Id             uuid.UUID `bson:"_id" json:"id"`
	WebsiteName    string    `bson:"websiteName" json:"websiteName"`
	Url            string    `bson:"url"`
	Content        string    `bson:"content"`
	HttpStatusCode byte      `bson:"httpStatusCode"`
	ContentSize    int       `bson:"contentSize"`
}
