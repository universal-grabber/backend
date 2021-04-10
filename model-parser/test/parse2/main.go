package main

import (
	"backend/model-parser/lib"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	jsonFile, err := os.Open("dev/payload_light.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	timeCalc := new(lib.TimeCalc)
	timeCalc.Init("T")

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	process(byteValue)
}

func process(p []byte) {
	resp, err := http.Post("https://127.0.0.1:8080/api/1.0/parse-light", "application/json", bytes.NewReader(p))

	data, err := ioutil.ReadAll(resp.Body)

	log.Print(string(data))

	if err != nil {
		log.Panic(err)
	}
}
