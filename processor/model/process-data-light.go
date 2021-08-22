package model

import "backend/gen/proto/base"

type ProcessDataLight struct {
	Html    *string       `json:"html" binding:"required"`
	PageRef *base.PageRef `json:"pageRef" binding:"required"`
}
