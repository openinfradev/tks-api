package helper

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/log"
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

func Transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(in); err != nil {
		log.Error(err)
	}
	if err := json.NewDecoder(buf).Decode(out); err != nil {
		log.Error(err)
	}
}

func GenerateEmailCode() string {
	num, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		panic(err)
	}
	num = num.Add(num, big.NewInt(100000))
	numString := fmt.Sprintf("%06d", num)
	return fmt.Sprintf("%06s", numString)
}
