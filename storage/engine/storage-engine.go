package engine

import (
	"backend/storage/engine/mongo-old"
	mongoStore "backend/storage/engine/mongo-store"
	uuid "github.com/satori/go.uuid"
)

type StorageEngineBackend interface {
	Get(uuid uuid.UUID) ([]byte, error)

	Delete(uuid uuid.UUID) (bool, error)

	Exists(uuid uuid.UUID) (bool, error)

	Add(uuid uuid.UUID, data []byte) error
}

var engines = map[string]StorageEngineBackend{
	"default":     mongo_old.NewInstance("ugr", "pageHtml"),
	"mongo-store": mongoStore.NewInstance("pageData", "html"),
}

func GetEngineBackendByClassName(engineClass string) StorageEngineBackend {
	return engines[engineClass]
}
