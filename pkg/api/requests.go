package api

type UserRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
