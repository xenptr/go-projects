package pgx

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xenptr/go-projects/blogging-platform-api/internal/models"
	"github.com/xenptr/go-projects/blogging-platform-api/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const selectPost = `
SELECT
	id,
	title,
	content,
	category,
	tags,
	created_at,
	updated_at
`

// pgTypeMap is reused across all scan calls to avoid per-call allocation.
var pgTypeMap = pgtype.NewMap()

// Store implements store.PostStore using a pgxpool connection pool.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ── helpers

// validatePost checks that the required fields Title and Content are present
// and non-empty. Returns store.ErrInvalidInput on failure.
func validatePost(post models.Post) error {
	if post.Title == nil || strings.TrimSpace(*post.Title) == "" {
		return fmt.Errorf("%w: title is required", store.ErrInvalidInput)
	}
	if post.Content == nil || strings.TrimSpace(*post.Content) == "" {
		return fmt.Errorf("%w: content is required", store.ErrInvalidInput)
	}
	return nil
}

// scanPost scans one row into a models.Post.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanPost(row rowScanner) (models.Post, error) {
	var post models.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Category,
		pgTypeMap.SQLScanner(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	return post, err
}

// collectPosts drains pgx rows into a slice of posts.
func collectPosts(rows pgx.Rows) ([]models.Post, error) {
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// ── PostStore implementation

func (s *Store) AddPost(post models.Post) (int64, error) {
	if err := validatePost(post); err != nil {
		return 0, err
	}

	category := ""
	if post.Category != nil {
		category = *post.Category
	}

	var id int64
	err := s.pool.QueryRow(
		context.Background(),
		`INSERT INTO posts (title, content, category, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		*post.Title, *post.Content, category, post.Tags,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("addPost: %w", err)
	}

	return id, nil
}

func (s *Store) AllPosts() ([]models.Post, error) {
	rows, err := s.pool.Query(context.Background(), selectPost+`FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("allPosts: %w", err)
	}

	posts, err := collectPosts(rows)
	if err != nil {
		return nil, fmt.Errorf("allPosts: %w", err)
	}
	return posts, nil
}

func (s *Store) PostByID(id int64) (models.Post, error) {
	row := s.pool.QueryRow(
		context.Background(),
		selectPost+`FROM posts WHERE id = $1`, id,
	)

	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return post, store.ErrNotFound
		}
		return post, fmt.Errorf("postByID: %w", err)
	}
	return post, nil
}

func (s *Store) PostsByTerm(term string) ([]models.Post, error) {
	rows, err := s.pool.Query(
		context.Background(),
		selectPost+`FROM posts
		WHERE
			title       ILIKE '%' || $1 || '%'
			OR content  ILIKE '%' || $1 || '%'
			OR category ILIKE '%' || $1 || '%'
			OR EXISTS (
				SELECT 1 
				FROM unnest(tags) AS tag
				WHERE tag ILIKE '%' || $1 || '%'
			)`,
		term,
	)
	if err != nil {
		return nil, fmt.Errorf("postsByTerm: %w", err)
	}

	posts, err := collectPosts(rows)
	if err != nil {
		return nil, fmt.Errorf("postsByTerm: %w", err)
	}
	return posts, nil
}

func (s *Store) UpdatePost(id int64, post models.Post) error {
	if err := validatePost(post); err != nil {
		return err
	}

	setClauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	i := 1

	setClauses = append(setClauses, fmt.Sprintf("title = $%d", i))
	args = append(args, *post.Title)
	i++

	setClauses = append(setClauses, fmt.Sprintf("content = $%d", i))
	args = append(args, *post.Content)
	i++

	if post.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", i))
		args = append(args, *post.Category)
		i++
	}
	if post.Tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", i))
		args = append(args, post.Tags)
		i++
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE posts SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		i,
	)

	ct, err := s.pool.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("updatePost: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) DeletePost(id int64) error {
	ct, err := s.pool.Exec(context.Background(), "DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deletePost: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}
