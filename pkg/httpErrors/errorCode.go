package httpErrors

type ErrorCode string

var errorMap = map[ErrorCode]string{
	// Common
	"C_INVALID_ACCOUNT_ID":        "유효하지 않은 어카운트 아이디입니다. 어카운트 아이디를 확인하세요.",
	"C_INVALID_STACK_ID":          "유효하지 않은 스택 아이디입니다. 스택 아이디를 확인하세요.",
	"C_INVALID_CLUSTER_ID":        "유효하지 않은 클러스터 아이디입니다. 클러스터 아이디를 확인하세요.",
	"C_INVALID_APPGROUP_ID":       "유효하지 않은 앱그룹 아이디입니다. 앱그룹 아이디를 확인하세요.",
	"C_INVALID_ORGANIZATION_ID":   "유효하지 않은 조직 아이디입니다. 조직 아이디를 확인하세요.",
	"C_INVALID_CLOUD_ACCOUNT_ID":  "유효하지 않은 클라우드어카운트 아이디입니다. 클라우드어카운트 아이디를 확인하세요.",
	"C_INVALID_STACK_TEMPLATE_ID": "유효하지 않은 스택템플릿 아이디입니다. 스택템플릿 아이디를 확인하세요.",
	"C_INVALID_ASA_ID":            "유효하지 않은 앱서빙앱 아이디입니다. 앱서빙앱 아이디를 확인하세요.",
	"C_INVALID_ASA_TASK_ID":       "유효하지 않은 테스크 아이디입니다. 테스크 아이디를 확인하세요.",

	// Auth
	"A_INVALID_ID":              "아이디가 존재하지 않습니다.",
	"A_INVALID_PASSWORD":        "비밀번호가 일치하지 않습니다.",
	"A_SAME_OLD_PASSWORD":       "기존 비밀번호와 동일합니다.",
	"A_INVALID_TOKEN":           "사용자 토큰 오류",
	"A_INVALID_USER_CREDENTIAL": "비밀번호가 일치하지 않습니다.",
	"A_INVALID_ORIGIN_PASSWORD": "기존 비밀번호가 일치하지 않습니다.",
	"A_INVALID_CODE":            "인증번호가 일치하지 않습니다.",
	"A_NO_SESSION":              "세션 정보를 찾을 수 없습니다.",
	"A_EXPIRED_CODE":            "인증번호가 만료되었습니다.",

	// User
	"U_NO_USER": "해당 사용자 정보를 찾을 수 없습니다.",

	// CloudAccount
	"CA_INVALID_CLIENT_TOKEN_ID":    "유효하지 않은 토큰입니다. AccessKeyId, SecretAccessKey, SessionToken 을 확인후 다시 입력하세요.",
	"CA_INVALID_CLOUD_ACCOUNT_NAME": "유효하지 않은 클라우드계정 이름입니다. 클라우드계정 이름을 확인하세요.",

	// Dashboard
	"D_INVALID_CHART_TYPE":    "유효하지 않은 차트타입입니다.",
	"D_INVALID_PRIMARY_STACK": "프라이머리 스택이 정상적으로 설치되지 않았습니다. 스택을 확인하세요.",
	"D_NOT_FOUND_CHART":       "요청한 차트를 불러올 수 없습니다.",
	"D_NO_STACK":              "",

	// AppServeApp
	"D_NO_ASA": "요청한 앱아이디에 해당하는 어플리케이션이 없습니다.",

	// Stack
	"S_INVALID_STACK_TEMPLATE":      "스택 템플릿을 가져올 수 없습니다.",
	"S_INVALID_CLOUD_ACCOUNT":       "클라우드 계정설정을 가져올 수 없습니다.",
	"S_INVALID_STACK_NAME":          "유효하지 않은 스택 이름입니다. 스택 이름을 확인하세요.",
	"S_FAILED_FETCH_CLUSTERS":       "조직에 해당하는 클러스터를 가져오는데 실패했습니다.",
	"S_FAILED_FETCH_CLUSTER":        "클러스터를 가져오는데 실패했습니다.",
	"S_FAILED_FETCH_ORGANIZATION":   "조직 ID에 해당하는 조직을 가져오는데 실패했습니다.",
	"S_CREATE_ALREADY_EXISTED_NAME": "조직에 이미 존재하는 이름입니다.",
	"S_FAILED_TO_CALL_WORKFLOW":     "스택 생성에 실패하였습니다. 관리자에게 문의하세요.",
	"S_REMAIN_CLUSTER_FOR_DELETION": "프라이머리 클러스터를 지우기 위해서는 조직내의 모든 클러스터를 삭제해야 합니다.",
	"S_FAILED_GET_CLUSTERS":         "클러스터를 가져오는데 실패했습니다.",
	"S_FAILED_DELETE_EXISTED_ASA":   "지우고자 하는 스택에 남아 있는 앱서빙앱이 있습니다.",

	// Alert
	"AL_NOT_FOUND_ALERT": "지정한 앨럿이 존재하지 않습니다.",

	// AppGroup
	"AG_NOT_FOUND_CLUSTER":         "지장한 클러스터가 존재하지 않습니다.",
	"AG_NOT_FOUND_APPGROUP":        "지장한 앱그룹이 존재하지 않습니다.",
	"AG_FAILED_TO_CREATE_APPGROUP": "앱그룹 생성에 실패하였습니다.",
	"AG_FAILED_TO_CALL_WORKFLOW":   "워크플로우 호출에 실패하였습니다.",
}

func (m ErrorCode) GetText() string {
	if v, ok := errorMap[m]; ok {
		return v
	}
	return ""
}
