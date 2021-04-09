package client

import (
	"backend/processor/lib"
	"backend/processor/model"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type BackendStorageClient struct {
	config model.Config
}

func (client *BackendStorageClient) Init(config model.Config) {
	client.config = config
}

func (client *BackendStorageClient) Store(item *model.PageRef) *model.StoreResult {
	data, err := json.Marshal(item)

	lib.Check(err)

	resp, err := http.Post(client.config.UgbStorageUri+"/api/1.0/store", "application/json", bytes.NewReader(data))

	lib.Check(err)

	respBytes, err := ioutil.ReadAll(resp.Body)

	lib.Check(err)

	storeResult := new(model.StoreResult)

	err = json.Unmarshal(respBytes, storeResult)

	lib.Check(err)

	return storeResult
}

func (client *BackendStorageClient) Get(item *model.PageRef) *model.StoreResult {
	data, err := json.Marshal(item)

	lib.Check(err)

	resp, err := http.Post(client.config.UgbStorageUri+"/api/1.0/get", "application/json", bytes.NewReader(data))

	lib.Check(err)

	respBytes, err := ioutil.ReadAll(resp.Body)

	lib.Check(err)

	storeResult := new(model.StoreResult)

	err = json.Unmarshal(respBytes, storeResult)

	lib.Check(err)

	return storeResult
}
