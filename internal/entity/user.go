package entity

import (
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	BasicEntity
	Avatar   *string `json:"avatar"     bun:"avatar"`
	Login    *string `json:"login"   bun:"login"`
	Password *string `json:"password"   bun:"password"`
	FullName *string `json:"full_name"  bun:"full_name"`
	Status   *bool   `json:"status"     bun:"status"`
	Phone    *string `json:"phone"      bun:"phone"`
	Role     *string `json:"role"       bun:"role"`
}
