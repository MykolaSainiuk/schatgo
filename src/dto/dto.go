package dto

// RegisterInputDto
type RegisterInputDto struct {
	Name      string `json:"name" binding:"required,omitempty" validate:"required,min=2"`
	Password  string `json:"password" binding:"required,omitempty" validate:"required,min=6"`
	AvatarUri string `json:"avatarUri" binding:"required" validate:"datauri"`
}

// RegisterResponseDto
type RegisterResponseDto struct {
	UserId string `json:"userId"`
}

// LoginInputDto
type LoginInputDto struct {
	Name     string `json:"name" binding:"required,omitempty" validate:"required,min=2"`
	Password string `json:"password" binding:"required,min=6" validate:"required,min=6"`
}

// LoginOutputDto
type LoginOutputDto struct {
	AccessToken string `json:"access_token" binding:"required" validate:"required"`
	// RefreshToken string `json:"refresh_token" binding:"required"`
}
