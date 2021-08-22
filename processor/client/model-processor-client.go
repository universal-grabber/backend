package client

import (
	"backend/gen/proto/base"
	"backend/processor/model"
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type ModelProcessorClient struct {
	config model.Config
}

func (client *ModelProcessorClient) Init(config model.Config) {
	client.config = config
}

func (client *ModelProcessorClient) Parse(result string, pageRef *base.PageRef) (*model.Record, error) {
	processorData := model.ProcessDataLight{
		Html:    &result,
		PageRef: pageRef,
	}

	data, err := json.Marshal(processorData)

	if err != nil {
		return nil, err
	}

	res, err := http.Post(client.config.UgbModelProcessorUri+"/api/1.0/parse-light", "application/json", bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	resBytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	record := new(model.Record)

	err = json.Unmarshal(resBytes, record)

	if res.StatusCode != 200 {
		log.Error("unable to parse", string(resBytes))
		return nil, err
	}

	return record, err
}
