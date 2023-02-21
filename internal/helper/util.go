package helper

import "github.com/google/uuid"

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func nilUUID() uuid.UUID {
	nilId, _ := uuid.Parse("")
	return nilId
}
