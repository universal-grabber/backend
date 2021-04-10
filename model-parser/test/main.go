package main

//
//import (
//	"encoding/json"
//	"fmt"
//	"io/ioutil"
//	"backend/model-parser/lib"
//	"backend/model-parser/model"
//	"os"
//)
//
//func main() {
//	p := model.ProcessData{}
//	jsonFile, err := os.Open("dev/payload.json")
//	if err != nil {
//		fmt.Println(err)
//	}
//	defer jsonFile.Close()
//
//	byteValue, _ := ioutil.ReadAll(jsonFile)
//	json.Unmarshal(byteValue, &p)
//
//	processor := new(lib.Processor)
//
//	timeCalc := new(lib.TimeCalc)
//	timeCalc.Init("T")
//
//	for i := 0; i < 8; i++ {
//		go func() {
//			for {
//				processor.ProcessData(p)
//				timeCalc.Step()
//			}
//		}()
//	}
//
//	for {
//		processor.ProcessData(p)
//		timeCalc.Step()
//	}
//}
