package helper

import (
	"math/rand"
	"strings"
	"time"
)

var (
	TKS_ADMIN                   = "tks-admin"
	PREFIX_USER_ID              = "u"
	PREFIX_CLUSTER_ID           = "c"
	PREFIX_ORGANIZATION_ID      = "o"
	PREFIX_APPLICATION_GROUP_ID = "a"
	ID_LENGTH                   = 9
)

const LETTERS_FOR_ID = "abcdefghijklmnopqrstuvwxyz0123456789" // lowercase RFC 1123

func GenerateClusterId() string {
	return PREFIX_CLUSTER_ID + randomString(ID_LENGTH-1)
}

func GenerateOrganizationId() string {
	return PREFIX_ORGANIZATION_ID + randomString(ID_LENGTH-1)
}

func GenerateApplicaionGroupId() string {
	return PREFIX_APPLICATION_GROUP_ID + randomString(ID_LENGTH-1)
}

func ValidateClusterId(id string) bool {
	if id == TKS_ADMIN {
		return true
	}
	if !strings.HasPrefix(id, PREFIX_CLUSTER_ID) {
		return false
	}
	return validateId(id)
}

func ValidateOrganizationId(id string) bool {
	if !strings.HasPrefix(id, PREFIX_ORGANIZATION_ID) {
		return false
	}
	return validateId(id)
}

func ValidateApplicationGroupId(id string) bool {
	if !strings.HasPrefix(id, PREFIX_APPLICATION_GROUP_ID) {
		return false
	}
	return validateId(id)
}

func validateId(id string) bool {
	for i := 0; i < len(id); i++ {
		if !strings.Contains(LETTERS_FOR_ID, string(id[i])) {
			return false
		}
	}
	return len(id) == ID_LENGTH
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		b[i] = LETTERS_FOR_ID[r.Int63()%int64(len(LETTERS_FOR_ID))]
	}
	return string(b)
}
