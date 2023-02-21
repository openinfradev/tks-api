package domain

type User = struct {
	Id         string `json:"id"`
	AccountId  string `json:"accountId"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	Token      string `json:"token"`
	Authorized bool   `json:"authorized"`
}
