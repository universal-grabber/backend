package lib

import "backend/processor/model"

type ConfigProvider interface {
	GetConfig() model.Config
}
