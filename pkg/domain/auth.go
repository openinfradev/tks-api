package domain

type LoginRequest struct {
	AccountId      string `json:"accountId" validate:"required"`
	Password       string `json:"password" validate:"required"`
	OrganizationId string `json:"organizationId" validate:"required"`
}

type LoginResponse struct {
	User struct {
		AccountId    string       `json:"accountId"`
		Name         string       `json:"name"`
		Token        string       `json:"token"`
		Role         Role         `json:"role"`
		Organization Organization `json:"organization"`
	} `json:"user"`
}

type VerifyIdentityForLostIdRequest struct {
	OrganizationId string `json:"organizationId" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	UserName       string `json:"userName" validate:"required"`
}

type FindIdRequest struct {
	OrganizationId string `json:"organizationId" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	UserName       string `json:"userName" validate:"required"`
	Code           string `json:"code" validate:"required"`
}

type FindIdResponse struct {
	AccountId string `json:"accountId"`
}

type VerifyIdentityForLostPasswordRequest struct {
	OrganizationId string `json:"organizationId" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	UserName       string `json:"userName" validate:"required"`
	AccountId      string `json:"accountId" validate:"required"`
}

type FindPasswordRequest struct {
	OrganizationId string `json:"organizationId" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	UserName       string `json:"userName" validate:"required"`
	AccountId      string `json:"accountId" validate:"required"`
	Code           string `json:"code" validate:"required"`
}
