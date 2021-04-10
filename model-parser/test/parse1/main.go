package main

import (
	"backend/model-parser/lib"
	"backend/model-parser/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	p := model.ProcessData{}

	jsonFile, err := os.Open("dev/payload.json")

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &p)

	parser := new(lib.Parser)

	result := parser.Parse(p)

	res, err := json.Marshal(result)

	err = ioutil.WriteFile("dev/result.json", res, os.ModePerm)
}
