package model

type WebSite struct {
	Id         string `bson:"_id" json:"id"`
	Name       string `json:"name"`
	Url        string `json:"url"`
	BatchDelay int    `json:"batch_delay"`
	BatchSize  int    `json:"batch_size"`
}
