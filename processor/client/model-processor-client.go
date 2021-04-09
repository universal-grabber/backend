package client

import (
	"backend/processor/lib"
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

func (client *ModelProcessorClient) Parse(result string, url string) *model.Record {
	processorData := model.ProcessDataLight{
		Html: &result,
		Url:  &url,
	}

	data, err := json.Marshal(processorData)

	lib.Check(err)

	res, err := http.Post(client.config.UgbModelProcessorUri+"/api/1.0/parse-light", "application/json", bytes.NewReader(data))

	lib.Check(err)

	resBytes, err := ioutil.ReadAll(res.Body)

	lib.Check(err)

	record := new(model.Record)

	err = json.Unmarshal(resBytes, record)

	if res.StatusCode != 200 {
		log.Panic("unable to parse", string(resBytes))
	}

	lib.Check(err)

	return record
}
