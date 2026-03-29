package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"wb_school/L3/L3.3/backend/internal/domain"
)

type CommentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	var path string
	var level int

	if comment.ParentID == nil {
		path = fmt.Sprintf("'%s'", comment.ID)
		level = 0
	} else {
		var parentPath string
		query := `SELECT path FROM comments WHERE id = $1 AND deleted_at IS NULL`
		err := r.db.GetContext(ctx, &parentPath, query, comment.ParentID)
		if err != nil {
			return err
		}
		path = fmt.Sprintf("'%s.%s'", parentPath, comment.ID)
		level = strings.Count(path, ".") - 1
	}

	comment.Path = path
	comment.Level = level

	query := `
		INSERT INTO comments (id, content, parent_id, author, created_at, updated_at, path, level)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		comment.ID, comment.Content, comment.ParentID, comment.Author,
		comment.CreatedAt, comment.UpdatedAt, comment.Path, comment.Level,
	)
	return err
}

func (r *CommentRepository) GetTree(ctx context.Context, parentID *string, limit, offset int, sortBy, order string) ([]*domain.CommentWithChildren, error) {
	query := `
		SELECT id, content, parent_id, author, created_at, updated_at, path, level
		FROM comments
		WHERE deleted_at IS NULL
		AND ($1::UUID IS NULL OR path <@ (SELECT path FROM comments WHERE id = $1))
		ORDER BY path
		LIMIT $2 OFFSET $3
	`

	var comments []*domain.Comment
	var err error

	if parentID == nil {
		err = r.db.SelectContext(ctx, &comments, query, nil, limit, offset)
	} else {
		err = r.db.SelectContext(ctx, &comments, query, parentID, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	commentMap := make(map[string]*domain.CommentWithChildren)
	var roots []*domain.CommentWithChildren

	for _, c := range comments {
		node := &domain.CommentWithChildren{
			Comment:  *c,
			Children: []*domain.CommentWithChildren{},
		}
		commentMap[c.ID] = node

		if c.ParentID == nil {
			roots = append(roots, node)
		} else {
			if parent, ok := commentMap[*c.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots, nil
}

func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE comments 
		SET deleted_at = NOW() 
		WHERE id = $1 OR path <@ (SELECT path FROM comments WHERE id = $1)
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *CommentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Comment, error) {
	var comments []*domain.Comment

	sqlQuery := `
		SELECT id, content, parent_id, author, created_at, updated_at, path, level
		FROM comments
		WHERE deleted_at IS NULL
		AND (to_tsvector('russian', content) @@ plainto_tsquery('russian', $1)
			OR content ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &comments, sqlQuery, query, limit, offset)
	return comments, err
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*domain.Comment, error) {
	var comment domain.Comment
	query := `SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &comment, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &comment, err
}
