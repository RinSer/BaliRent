package api


type UserRequest struct {
	Id       *int64 `json:"id,omitempty"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
