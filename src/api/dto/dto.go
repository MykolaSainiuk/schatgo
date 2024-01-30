package dto

// -- RegisterUser
// RegisterInputDto
type RegisterInputDto struct {
	Name      string `json:"name" validate:"required,min=2"`
	Password  string `json:"password" validate:"required,min=6"`
	AvatarUri string `json:"avatarUri" validate:"url|uri|base64url"`
}

// RegisterResponseDto
type RegisterResponseDto struct {
	UserId string `json:"id"`
}

// -- LoginUser
// LoginInputDto
type LoginInputDto struct {
	Name     string `json:"name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginOutputDto
type LoginOutputDto struct {
	AccessToken string `json:"token"`
	// RefreshToken string `json:"refresh_token"`
}

// UserInfoOutputDto
type UserInfoOutputDto struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarUri string `json:"avatarUri"`
}

type AddContactInputDto struct {
	UserName string `json:"username" validate:"required,min=2"`
}
