package httpErrors

type ErrorCode string

var errorMap = map[ErrorCode]string{
	"INVALID_CLIENT_TOKEN_ID": "유효하지 않은 토큰입니다. AccessKeyId, SecretAccessKey, SessionToken 을 확인후 다시 입력하세요.",
}

func (m ErrorCode) GetText() string {
	if v, ok := errorMap[m]; ok {
		return v
	}
	return ""
}
