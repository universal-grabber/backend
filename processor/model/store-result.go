package model

type StoreResult struct {
	Content string `json:"content"`
	Size    int    `json:"size"`
	Ok      bool   `json:"ok"`
}
