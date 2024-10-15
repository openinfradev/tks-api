package helper

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func GenerateEmailCode(ctx context.Context) (string, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		log.Error(ctx, err)
		return "", err
	}
	num = num.Add(num, big.NewInt(100000))
	numString := fmt.Sprintf("%06d", num)
	return fmt.Sprintf("%06s", numString), nil
}

func IsDurationExpired(targetTime time.Time, duration time.Duration) bool {
	now := time.Now()
	diff := now.Sub(targetTime)
	return diff > duration
}

func SplitAddress(ctx context.Context, url string) (address string, port int) {
	url = strings.TrimSuffix(url, "\n")
	arr := strings.Split(url, ":")
	address = arr[0] + ":" + arr[1]

	portNum := 80
	if len(arr) == 3 {
		portNum, _ = strconv.Atoi(arr[2])
	} else {
		if strings.Contains(arr[0], "https") {
			portNum = 443
		}
	}
	port = portNum

	log.Infof(ctx, "address : %s, port : %d", address, port)
	return
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func BoolP(value bool) *bool {
	return &value
}

func StringP(value string) *string {
	return &value
}

func UUIDP(value uuid.UUID) *uuid.UUID {
	return &value
}

func DeepCopy(src, dest interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, dest)
}
