package user

import (
	"github.com/uptrace/bun"
	"mime/multipart"
	"time"
)

type Filter struct {
	Limit  *int
	Offset *int
	Page   *int
	Search *string
	Role   *string
}

type SignInRequest struct {
	Login    string `json:"login" form:"login"`
	Password string `json:"password" form:"password"`
}

type AuthClaims struct {
	ID   int
	Role string
}

type GetListResponse struct {
	ID       int     `json:"id"`
	Avatar   *string `json:"avatar"`
	FullName *string `json:"full_name"`
	Login    *string `json:"login"`
	Status   *bool   `json:"status"`
	Phone    *string `json:"phone"`
	Role     *string `json:"role"`
}

type GetDetailByIdResponse struct {
	ID       int     `json:"id"`
	Avatar   *string `json:"avatar"`
	Login    *string `json:"login"`
	FullName *string `json:"full_name"`
	Status   *bool   `json:"status"`
	Phone    *string `json:"phone"`
	Role     *string `json:"role"`
}

type CreateRequest struct {
	Login      *string               `json:"login" form:"login"`
	FullName   *string               `json:"full_name" form:"full_name"`
	Password   *string               `json:"password" form:"password"`
	Phone      *string               `json:"phone" form:"phone"`
	Avatar     *multipart.FileHeader `json:"-" form:"avatar"`
	AvatarLink *string               `json:"-" form:"-"`
	Role       *string               `json:"role" form:"role"`
}

type CreateResponse struct {
	bun.BaseModel `bun:"table:users"`

	ID        int       `json:"id" bun:"-"`
	Avatar    *string   `json:"avatar"     bun:"avatar"`
	Login     *string   `json:"login"   bun:"login"`
	Password  *string   `json:"-"   bun:"password"`
	FullName  *string   `json:"full_name" bun:"full_name"`
	Phone     *string   `json:"phone"      bun:"phone"`
	Role      *string   `json:"role" bun:"role"`
	CreatedAt time.Time `json:"-"          bun:"created_at"`
	CreatedBy int       `json:"-"          bun:"created_by"`
}

type UpdateRequest struct {
	ID         int                   `json:"id" form:"id"`
	Login      *string               `json:"login" form:"login"`
	FullName   *string               `json:"full_name" form:"full_name"`
	Password   *string               `json:"password" form:"password"`
	Phone      *string               `json:"phone" form:"phone"`
	Avatar     *multipart.FileHeader `json:"-" form:"avatar"`
	AvatarLink *string               `json:"-" form:"-"`
	Role       *string               `json:"role" form:"role"`
}

type UploadAvatarRequest struct {
	ID         int                   `json:"id" form:"id"`
	Avatar     *multipart.FileHeader `json:"-" form:"avatar"`
	AvatarLink *string               `json:"-" form:"-"`
}

type GetMeResponse struct {
	ID       int     `json:"id"`
	Login    *string `json:"login"`
	Avatar   *string `json:"avatar"`
	FullName *string `json:"full_name"`
	Phone    *string `json:"phone"`
}
