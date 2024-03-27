package httpErrors

type ErrorCode string

var errorMap = map[ErrorCode]string{
	// Common
	"C_INTERNAL_ERROR":                          "예상하지 못한 오류가 발생했습니다. 문제가 계속되면 관리자에게 문의해주세요.",
	"C_INVALID_ACCOUNT_ID":                      "유효하지 않은 어카운트 아이디입니다. 어카운트 아이디를 확인하세요.",
	"C_INVALID_STACK_ID":                        "유효하지 않은 스택 아이디입니다. 스택 아이디를 확인하세요.",
	"C_INVALID_CLUSTER_ID":                      "유효하지 않은 클러스터 아이디입니다. 클러스터 아이디를 확인하세요.",
	"C_INVALID_APPGROUP_ID":                     "유효하지 않은 앱그룹 아이디입니다. 앱그룹 아이디를 확인하세요.",
	"C_INVALID_ORGANIZATION_ID":                 "유효하지 않은 조직 아이디입니다. 조직 아이디를 확인하세요.",
	"C_INVALID_PROJECT_ID":                      "유효하지 않은 프로젝트 아이디입니다. 아이디를 확인하세요.",
	"C_INVALID_CLOUD_ACCOUNT_ID":                "유효하지 않은 클라우드어카운트 아이디입니다. 클라우드어카운트 아이디를 확인하세요.",
	"C_INVALID_STACK_TEMPLATE_ID":               "유효하지 않은 스택템플릿 아이디입니다. 스택템플릿 아이디를 확인하세요.",
	"C_INVALID_SYSTEM_NOTIFICATION_TEMPLATE_ID": "유효하지 않은 알림템플릿 아이디입니다. 알림템플릿 아이디를 확인하세요.",
	"C_INVALID_SYSTEM_NOTIFICATION_RULE_ID":     "유효하지 않은 알림설정 아이디입니다. 알림설정 아이디를 확인하세요.",
	"C_INVALID_ASA_ID":                          "유효하지 않은 앱서빙앱 아이디입니다. 앱서빙앱 아이디를 확인하세요.",
	"C_INVALID_ASA_TASK_ID":                     "유효하지 않은 테스크 아이디입니다. 테스크 아이디를 확인하세요.",
	"C_INVALID_CLOUD_SERVICE":                   "유효하지 않은 클라우드서비스입니다.",
	"C_INVALID_AUDIT_ID":                        "유효하지 않은 로그 아이디입니다. 로그 아이디를 확인하세요.",
	"C_INVALID_POLICY_TEMPLATE_ID":              "유효하지 않은 정책 템플릿 아이디입니다. 정책 템플릿 아이디를 확인하세요.",
	"C_INVALID_POLICY_ID":                       "유효하지 않은 정책 아이디입니다. 정책 아이디를 확인하세요.",
	"C_FAILED_TO_CALL_WORKFLOW":                 "워크플로우 호출에 실패했습니다.",

	// Auth
	"A_INVALID_ID":              "아이디가 존재하지 않습니다.",
	"A_INVALID_PASSWORD":        "비밀번호가 일치하지 않습니다.",
	"A_SAME_OLD_PASSWORD":       "기존 비밀번호와 동일합니다.",
	"A_INVALID_TOKEN":           "사용자 토큰 오류",
	"A_EXPIRED_TOKEN":           "사용자 토큰 만료",
	"A_INVALID_USER_CREDENTIAL": "비밀번호가 일치하지 않습니다.",
	"A_INVALID_ORIGIN_PASSWORD": "기존 비밀번호가 일치하지 않습니다.",
	"A_INVALID_CODE":            "인증번호가 일치하지 않습니다.",
	"A_NO_SESSION":              "세션 정보를 찾을 수 없습니다.",
	"A_EXPIRED_CODE":            "인증번호가 만료되었습니다.",
	"A_UNUSABLE_TOKEN":          "사용할 수 없는 토큰입니다.",

	// Organization
	"O_INVALID_ORGANIZATION_NAME":                   "조직에 이미 존재하는 이름입니다.",
	"O_NOT_EXISTED_NAME":                            "조직이 존재하지 않습니다.",
	"O_FAILED_UPDATE_STACK_TEMPLATES":               "조직에 스택템플릿을 설정하는데 실패했습니다",
	"O_FAILED_UPDATE_POLICY_TEMPLATES":              "조직에 정책템플릿을 설정하는데 실패했습니다",
	"O_FAILED_UPDATE_SYSTEM_NOTIFICATION_TEMPLATES": "조직에 알림템플릿을 설정하는데 실패했습니다",

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

	// Cluster
	"CL_INVALID_BYOH_CLUSTER_ENDPOINT": "BYOH 타입의 클러스터 생성을 위한 cluster endpoint 가 유효하지 않습니다.",
	"CL_INVALID_CLUSTER_TYPE_AWS":      "클러스터 타입이 유효하지 않습니다.",

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
	"S_NOT_ENOUGH_QUOTA":            "AWS 의 resource quota 가 부족합니다. 관리자에게 문의하세요.",
	"S_INVALID_CLUSTER_URL":         "BYOH 타입의 클러스터 생성은 반드시 userClusterEndpoint 값이 필요합니다.",
	"S_INVALID_CLUSTER_ID":          "BYOH 타입의 클러스터 생성은 반드시 clusterId 값이 필요합니다.",
	"S_INVALID_CLOUD_SERVICE":       "클라우드 서비스 타입이 잘못되었습니다.",

	// Alert
	"AL_NOT_FOUND_ALERT": "지정한 앨럿이 존재하지 않습니다.",

	// SystemNotificationTemplate
	"SNT_CREATE_ALREADY_EXISTED_NAME": "알림템플릿에 이미 존재하는 이름입니다.",
	"SNT_FAILED_FETCH_ALERT_TEMPLATE": "알림템플릿을 가져오는데 실패했습니다.",
	"SNT_FAILED_UPDATE_ORGANIZATION":  "알림템플릿에 조직을 설정하는데 실패했습니다.",
	"SNT_NOT_EXISTED_ALERT_TEMPLATE":  "업데이트할 알림템플릿이 존재하지 않습니다.",
	"SNT_FAILED_DELETE_EXIST_RULES":   "알림템플릿을 사용하고 있는 알림 설정이 있습니다. 알림 설정을 삭제하세요.",

	// SystemNotificationRule
	"SNR_CREATE_ALREADY_EXISTED_NAME":           "알림 설정에 이미 존재하는 이름입니다.",
	"SNR_FAILED_FETCH_SYSTEM_NOTIFICATION_RULE": "알림 설정을 가져오는데 실패했습니다.",
	"SNR_FAILED_UPDATE_ORGANIZATION":            "알림 설정에 조직을 설정하는데 실패했습니다.",
	"SNR_NOT_EXISTED_SYSTEM_NOTIFICATION_RULE":  "업데이트할 알림 설정이 존재하지 않습니다.",

	// AppGroup
	"AG_NOT_FOUND_CLUSTER":         "지장한 클러스터가 존재하지 않습니다.",
	"AG_NOT_FOUND_APPGROUP":        "지장한 앱그룹이 존재하지 않습니다.",
	"AG_FAILED_TO_CREATE_APPGROUP": "앱그룹 생성에 실패하였습니다.",
	"AG_FAILED_TO_CALL_WORKFLOW":   "워크플로우 호출에 실패하였습니다.",

	// StackTemplate
	"ST_CREATE_ALREADY_EXISTED_NAME":                             "스택템플릿에 이미 존재하는 이름입니다.",
	"ST_FAILED_UPDATE_ORGANIZATION":                              "스택템플릿에 조직을 설정하는데 실패했습니다.",
	"ST_NOT_EXISTED_STACK_TEMPLATE":                              "업데이트할 스택템플릿이 존재하지 않습니다.",
	"ST_INVALID_STACK_TEMAPLTE_NAME":                             "유효하지 않은 스택템플릿 이름입니다. 스택템플릿 이름을 확인하세요.",
	"ST_FAILED_FETCH_STACK_TEMPLATE":                             "스택템플릿을 가져오는데 실패했습니다.",
	"ST_FAILED_ADD_ORGANIZATION_STACK_TEMPLATE":                  "조직에 스택템플릿을 추가하는데 실패하였습니다.",
	"ST_FAILED_REMOVE_ORGANIZATION_STACK_TEMPLATE":               "조직에서 스택템플릿을 삭제하는데 실패하였습니다.",
	"ST_FAILED_ADD_ORGANIZATION_SYSTEM_NOTIFICATION_TEMPLATE":    "조직에 시스템알람템플릿을 추가하는데 실패하였습니다.",
	"ST_FAILED_REMOVE_ORGANIZATION_SYSTEM_NOTIFICATION_TEMPLATE": "조직에서 시스템알람템플릿을 삭제하는데 실패하였습니다.",
	"ST_FAILED_DELETE_EXIST_CLUSTERS":                            "스택템플릿을 사용하고 있는 스택이 있습니다. 스택을 삭제하세요.",

	// PolicyTemplate
	"PT_CREATE_ALREADY_EXISTED_NAME":       "정첵 템플릿에 이미 존재하는 이름입니다.",
	"PT_CREATE_ALREADY_EXISTED_KIND":       "정책 템플릿에 이미 존재하는 유형입니다.",
	"PT_NOT_FOUND_POLICY_TEMPLATE":         "정책 템플릿이 존재하지 않습니다.",
	"PT_INVALID_KIND":                      "유효하지 않은 정책 템플릿 유형입니다. 정책 템플릿 유형을 확인하세요.",
	"PT_FAILED_FETCH_POLICY_TEMPLATE":      "정책 템플릿 ID에 해당하는 정책 템플릿을 가져오는데 실패했습니다.",
	"PT_INVALID_REGO_SYNTAX":               "Rego 문법 오류입니다.",
	"PT_INVALID_POLICY_TEMPLATE_VERSION":   "유효하지 않은 정책 템플릿 버전닙니다. 정책 템플릿 버전을 확인하세요.",
	"PT_NOT_FOUND_POLICY_TEMPLATE_VERSION": "정책 템플릿 버전이 존재하지 않습니다.",
	"PT_INVALID_POLICY_TEMPLATE_NAME":      "유효하지 않은 정책 템플릿 이름입니다. 정책 템플릿 이름을 확인하세요.",
	"PT_INVALID_POLICY_TEMPLATE_KIND":      "유효하지 않은 정책 템플릿 유형입니다. 정책 템플릿 유형을 확인하세요.",
	"PT_INVALID_REGO_PARSEPARAMETER":       "유효하지 않은 Rego 파싱 설정입니다. Rego 파싱 설정을 확인하세요.",

	// Policy
	"P_CREATE_ALREADY_EXISTED_NAME": "정첵에 이미 존재하는 이름입니다.",
	"P_NOT_FOUND_POLICY":            "정책이 존재하지 않습니다.",
	"P_INVALID_POLICY_NAME":         "유효하지 않은 정책 이름입니다. 정책 이름을 확인하세요.",
}

func (m ErrorCode) GetText() string {
	if v, ok := errorMap[m]; ok {
		return v
	}
	return ""
}
