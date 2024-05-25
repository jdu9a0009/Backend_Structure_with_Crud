package entity

import "time"

type BasicEntity struct {
	ID        int        `json:"id" bun:"id"`
	CreatedAt *time.Time `json:"created_at"          bun:"created_at"`
	CreatedBy *int       `json:"created_by"          bun:"created_by"`
	UpdatedAt *time.Time `json:"updated_at"          bun:"updated_at"`
	UpdatedBy *int       `json:"updated_by"          bun:"updated_by"`
	DeletedAt *time.Time `json:"deleted_at"          bun:"deleted_at"`
	DeletedBy *int       `json:"deleted_by"          bun:"deleted_by"`
}
