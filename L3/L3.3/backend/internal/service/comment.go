package service

import (
	"context"
	"fmt"
	"strings"

	"wb_school/L3/L3.3/backend/internal/domain"
	"wb_school/L3/L3.3/backend/internal/infrastructure/postgres"
)

type CommentService struct {
	repo *postgres.CommentRepository
}

func NewCommentService(repo *postgres.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) Create(ctx context.Context, content, author string, parentID *string) (*domain.Comment, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	if strings.TrimSpace(author) == "" {
		return nil, fmt.Errorf("author cannot be empty")
	}

	comment := &domain.Comment{
		Content:  content,
		Author:   author,
		ParentID: parentID,
	}

	if parentID != nil {
		parent, err := s.repo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent comment not found")
		}
	}

	err := s.repo.Create(ctx, comment)
	return comment, err
}

func (s *CommentService) GetTree(ctx context.Context, parentID *string, page, pageSize int, sortBy, order string) ([]*domain.CommentWithChildren, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	tree, err := s.repo.GetTree(ctx, parentID, pageSize, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}

	// TODO: реализовать подсчет общего количества
	total := len(tree)

	return tree, total, nil
}

func (s *CommentService) Delete(ctx context.Context, id string) error {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if comment == nil {
		return fmt.Errorf("comment not found")
	}

	return s.repo.Delete(ctx, id)
}

func (s *CommentService) Search(ctx context.Context, query string, page, pageSize int) ([]*domain.Comment, int, error) {
	if strings.TrimSpace(query) == "" {
		return nil, 0, fmt.Errorf("search query cannot be empty")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	comments, err := s.repo.Search(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total := len(comments)

	return comments, total, nil
}
