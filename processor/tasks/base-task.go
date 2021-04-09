package tasks

import (
	"backend/processor/client"
)

type BaseTask interface {
	Init(clients client.Clients)
	Run()
	Name() string
}
