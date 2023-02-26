package idgenerator

import "github.com/segmentio/ksuid"

func GenerateId() string {
	return ksuid.New().String()
}
