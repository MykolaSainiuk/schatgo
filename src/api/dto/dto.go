package dto

//-- RegisterUser
// RegisterInputDto
type RegisterInputDto struct {
	Name      string `json:"name" binding:"required,omitempty" validate:"required,min=2"`
	Password  string `json:"password" binding:"required,omitempty" validate:"required,min=6"`
	AvatarUri string `json:"avatarUri" binding:"required" validate:"datauri"`
}
// RegisterResponseDto
type RegisterResponseDto struct {
	UserId string `json:"id"`
}

//-- LoginUser
// LoginInputDto
type LoginInputDto struct {
	Name     string `json:"name" binding:"required,omitempty" validate:"required,min=2"`
	Password string `json:"password" binding:"required,min=6" validate:"required,min=6"`
}
// LoginOutputDto
type LoginOutputDto struct {
	AccessToken string `json:"token"`
	// RefreshToken string `json:"refresh_token"`
}
// GetMeOutputDto
type GetMeOutputDto struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarUri string `json:"avatarUri"`
}
