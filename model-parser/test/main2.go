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
	jsonFile, err := os.Open("dev/payload.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	timeCalc := new(lib.TimeCalc)
	timeCalc.Init("T")

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for i := 0; i < 8; i++ {
		go func() {
			for {
				process(byteValue)
				timeCalc.Step()
			}
		}()
	}

	for {
		process(byteValue)
		timeCalc.Step()
	}
}

func process(p []byte) {
	_, err := http.Post("https://10.0.1.77:30004/api/1.0/parse", "application/json", bytes.NewReader(p))

	if err != nil {
		log.Panic(err)
	}
}
