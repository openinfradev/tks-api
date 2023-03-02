package helper

import (
	"encoding/json"

	"github.com/google/uuid"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func NilUUID() uuid.UUID {
	nilId, _ := uuid.Parse("")
	return nilId
}

func ModelToJson(in any) string {
	a, _ := json.Marshal(in)
	n := len(a)        //Find the length of the byte array
	s := string(a[:n]) //convert to string
	return s
}
