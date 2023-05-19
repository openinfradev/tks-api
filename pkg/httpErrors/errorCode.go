package httpErrors

type ErrorCode string

var errorMap = map[ErrorCode]string{
	"A_INVALID_USER_CREDENTIAL": "비밀번호가 일치하지 않습니다.",
	"A_INVALID_ORIGIN_PASSWORD": "기존 비밀번호가 일치하지 않습니다.",
	"A_MISMATCH_PASSWORD":       "비밀번호가 일치하지 않습니다.",
	"A_MISMATCH_CODE":           "인증번호가 일치하지 않습니다.",

	"CA_INVALID_CLIENT_TOKEN_ID": "유효하지 않은 토큰입니다. AccessKeyId, SecretAccessKey, SessionToken 을 확인후 다시 입력하세요.",

	"D_INVALID_PRIMARY_STACK": "프라이머리 스택이 정상적으로 설치되지 않았습니다. 스택을 확인하세요.",
	"D_NOT_FOUND_CHART":       "요청한 차트를 불러올 수 없습니다.",
	"D_NO_STACK":              "",
}

func (m ErrorCode) GetText() string {
	if v, ok := errorMap[m]; ok {
		return v
	}
	return ""
}
