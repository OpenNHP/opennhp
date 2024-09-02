package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/server"
)

func TestJsonTypeDetermine(t *testing.T) {
	jsonStr := "{\"data\":{\"accessToken\":\"abcdefgh\"}}"

	type DataStruct struct {
		Data map[string]any `json:"data"`
	}

	var data map[string]any

	obj := &DataStruct{}

	err := json.Unmarshal([]byte(jsonStr), obj)
	if err != nil {
		fmt.Println("unmarshal error: ", err)
		return
	}

	tokenStr, ok := obj.Data["accessToken"].(string)
	if !ok {
		fmt.Println("type conversion failed")
	}

	fmt.Println("token string: ", tokenStr)

	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("unmarshal1 error: ", err)
		return
	}

	tokenStr1, ok := obj.Data["accessToken"].(string)
	if !ok {
		fmt.Println("type conversion1 failed")
	}

	fmt.Println("token string1: ", tokenStr1)
}

func TestJsonSerialization(t *testing.T) {
	var data common.ACOpsResultMsg
	data.ErrCode = "40001"
	data.ErrMsg = "This is an error test"

	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("marshal data error: ", err)
	}

	fmt.Println("json string of data: ", string(bytes))

	var data2 map[string]string
	var dataStr string = "{\"errCode\":\"50001\", \"errMsg\":\"This is the second data.\", \"data\":{\"token\":\"654321\"}}"
	err = json.Unmarshal([]byte(dataStr), &data2)
	if err != nil {
		fmt.Println("unmarshal error: ", err)
	}
	fmt.Printf("json object: %+v\n", data2)
	bytes, err = json.Marshal(data2)
	if err != nil {
		fmt.Println("marshal data2 error: ", err)
	}

	fmt.Println("json string of data2: ", string(bytes))
}

type UserData struct {
	B0 bool
	B1 bool
	I0 int
	I1 int
}

func TestJsonSerializationForAnyType(t *testing.T) {
	otpMsg := &common.AgentOTPMsg{
		UserId:   "abc",
		UserData: map[string]any{"bool1": true, "bool2": false, "num1": 5555, "num2": 6666},
	}

	otpBytes, _ := json.Marshal(otpMsg)
	fmt.Println("json string of otpMsg: ", string(otpBytes))

	otpMsg1 := &common.AgentOTPMsg{}
	err := json.Unmarshal(otpBytes, otpMsg1)
	if err != nil {
		fmt.Println("unmarshal error: ", err)
	}
	fmt.Printf("json object: %+v\n", *otpMsg1)

	otpBytes1, _ := json.Marshal(otpMsg1)
	fmt.Println("json string of otpMsg1: ", string(otpBytes1))
}

func TestServerJsonConfig(t *testing.T) {
	fileName := filepath.Join("../server/main/etc", "config.json")
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Failed to load config file: ", err)
		return
	}

	config := &server.Config{}

	err = json.Unmarshal(content, config)
	if err != nil {
		fmt.Println("json format error: ", err)
		return
	}

	fmt.Printf("config: %+v\n", config)
}
