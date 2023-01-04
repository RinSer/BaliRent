package api

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
