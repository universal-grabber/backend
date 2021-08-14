package client

import "backend/processor/lib"

type Clients interface {
	lib.ConfigProvider
	GetApiClient() *ApiClient
	GetBackendStorageClient() *BackendStorageClient
	GetModelProcessorClient() *ModelProcessorClient
	GetPublisherClient() *PublisherClient
}
