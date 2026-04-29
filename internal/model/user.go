package model

// AuthUser представляет данные для аутентификации
type AuthUser struct {
	Login    string `json:"login" example:"user@example.com" validate:"required,min=3,max=100"`
	Password string `json:"password" example:"strongpassword123" validate:"required,min=6,max=100"`
}

// AuthResponse представляет ответ с токеном
type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UserDB представляет пользователя в базе данных
type UserDB struct {
	ID       int    `json:"id" example:"1"`
	Login    string `json:"login" example:"user@example.com"`
	Password string `json:"password" example:"hashed_password"`
}

type User struct {
	ID    int
	Login string
	Token string
}
