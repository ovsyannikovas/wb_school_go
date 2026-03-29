package domain

import (
	"time"
)

type Comment struct {
	ID        string     `json:"id" db:"id"`
	Content   string     `json:"content" db:"content"`
	ParentID  *string    `json:"parent_id,omitempty" db:"parent_id"`
	Author    string     `json:"author" db:"author"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Path      string     `json:"path" db:"path"`
	Level     int        `json:"level" db:"level"`
}

type CommentWithChildren struct {
	Comment
	Children []*CommentWithChildren `json:"children,omitempty"`
}
