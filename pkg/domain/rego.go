package domain

type ParameterDef struct {
	Key          string          `json:"key"`
	Type         string          `json:"type"`
	DefaultValue string          `json:"defaultValue"`
	Children     []*ParameterDef `json:"children"`
	IsArray      bool
}

type RegoCompileRequest struct {
	Rego string `json:"rego" example:"Rego 코드"`
}

type RegoCompieError struct {
	Status  int    `json:"status" example:"400"`
	Code    string `json:"code" example:"P_INVALID_REGO_SYNTAX"`
	Message string `json:"message" example:"Invalid rego syntax"`
	Text    string `json:"text" example:"Rego 문법 에러입니다. 라인:2 컬럼:1 에러메시지: var testnum is not safe"`
}

type RegoCompileResponse struct {
	ParametersSchema []*ParameterDef   `json:"parametersSchema,omitempty"`
	Errors           []RegoCompieError `json:"errors,omitempty"`
}
