package httpErrors

type ErrorCode string

var errorMap = map[ErrorCode]string{
	"INVALID_CLIENT_TOKEN_ID": "유효하지 않은 클라이언트 토큰 아이디입니다.",
}

func (m ErrorCode) GetText() string {
	if v, ok := errorMap[m]; ok {
		return v
	}
	return ""
}
