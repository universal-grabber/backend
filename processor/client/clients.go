package client

import "backend/processor/lib"

type Clients interface {
	lib.ConfigProvider
	GetApiClient() *ApiClientNew
	GetBackendStorageClient() *BackendStorageClient
	GetModelProcessorClient() *ModelProcessorClient
	GetPublisherClient() *PublisherClient
}
