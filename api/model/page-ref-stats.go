package model

type PageRefStats struct {
	Source string `bson:"websiteName" json:"websiteName"`
	State  string `json:"state"`
	Status string `json:"status"`
	Count  int64  `json:"count"`
}
