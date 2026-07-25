package store

import (
	"blogging-platform-api/internal/models"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
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

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) AddBlog(post models.Post) (int64, error) {
	if post.Title == nil || strings.TrimSpace(*post.Title) == "" {
		return 0, fmt.Errorf("title is required")
	}
	if post.Content == nil || strings.TrimSpace(*post.Content) == "" {
		return 0, fmt.Errorf("content is required")
	}

	var id int64

	category := ""
	if post.Category != nil {
		category = *post.Category
	}

	err := s.db.QueryRow(
		`INSERT INTO posts (title, content, category, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		*post.Title, *post.Content, category, post.Tags,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("addBlog: %w", err)
	}

	return id, nil
}

func (s *Store) AllPosts() ([]models.Post, error) {
	var posts []models.Post
	rows, err := s.db.Query(selectPost + `FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("allPosts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post

		if err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Category,
			pgtype.NewMap().SQLScanner(&post.Tags),
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("allPosts: %w", err)
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("allPosts: %w", err)
	}

	return posts, nil
}

func (s *Store) PostByID(id int64) (models.Post, error) {
	var post models.Post

	row := s.db.QueryRow(
		selectPost+`FROM posts
		WHERE id = $1`, id,
	)

	if err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Category,
		pgtype.NewMap().SQLScanner(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return post, ErrNotFound
		}
		return post, fmt.Errorf("postByID: %w", err)
	}

	return post, nil
}

func (s *Store) PostsByTerm(term string) ([]models.Post, error) {
	var posts []models.Post
	rows, err := s.db.Query(
		selectPost+`FROM posts
		WHERE 
			title ILIKE '%' || $1 || '%'
			OR content ILIKE '%' || $1 || '%'
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
	defer rows.Close()

	for rows.Next() {
		var post models.Post

		if err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Category,
			pgtype.NewMap().SQLScanner(&post.Tags),
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("postsByTerm: %w", err)
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("postsByTerm: %w", err)
	}

	return posts, nil
}

func (s *Store) UpdatePost(id int64, post models.Post) error {
	if post.Title == nil || strings.TrimSpace(*post.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if post.Content == nil || strings.TrimSpace(*post.Content) == "" {
		return fmt.Errorf("content is required")
	}

	setClauses := []string{}
	args := []any{}
	i := 1

	// Title and Content are guaranteed non-nil by the validation above.
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

	if len(setClauses) == 0 {
		return fmt.Errorf("updatePost: no fields to update")
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE posts SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "),
		i,
	)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("updatePost: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("updatePost: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("updatePost: no post with id %d", id)
	}

	return nil
}

func (s *Store) DeletePost(id int64) error {
	result, err := s.db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deletePost: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("deletePost: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
