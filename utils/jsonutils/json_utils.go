package jsonutils

import (
	"encoding/json"
	"log"
)

func ToJsonString(obj any) string {
	jsonString, err := json.Marshal(obj)
	if err != nil {
		log.Println("Can not parse json ", obj, err)
		return ""
	}
	return string(jsonString)
}

func MustToJsonString(obj any) (string, error) {
	jsonString, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}
