package store

import (
	"context"
	"database/sql"
	"fmt"

	"content-public-api/model"
)

type ContentStore struct {
	db *sql.DB
}

func NewContentStore(db *sql.DB) *ContentStore {
	return &ContentStore{db: db}
}

const itemsPerSection = 5

func (s *ContentStore) GetGroupedSections(ctx context.Context) (model.SectionsResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, COALESCE(title, ''), COALESCE(slug, ''),
		       COALESCE(short_description, ''), COALESCE(message, ''), section
		FROM content
		WHERE status = 'PUBLISHED'
		  AND section IS NOT NULL
		  AND CAST(section AS SIGNED) > 0
		ORDER BY CAST(section AS SIGNED) ASC, last_updated DESC`)
	if err != nil {
		return model.SectionsResponse{}, err
	}
	defer rows.Close()

	sectionIndex := map[int]int{}
	counts := map[int]int{}
	var sections []model.Section

	for rows.Next() {
		var (
			id               int64
			title            string
			slug             string
			shortDescription string
			message          string
			rawSection       sql.NullInt64
		)
		if err := rows.Scan(&id, &title, &slug, &shortDescription, &message, &rawSection); err != nil {
			return model.SectionsResponse{}, err
		}

		if !rawSection.Valid || rawSection.Int64 <= 0 {
			continue
		}
		secNum := int(rawSection.Int64)

		if counts[secNum] >= itemsPerSection {
			continue
		}

		idx, seen := sectionIndex[secNum]
		if !seen {
			sections = append(sections, model.Section{
				Name:  fmt.Sprintf("Sección %d", secNum),
				Items: []model.SectionItem{},
			})
			idx = len(sections) - 1
			sectionIndex[secNum] = idx
		}

		sections[idx].Items = append(sections[idx].Items, model.SectionItem{
			ID:               id,
			Title:            title,
			Slug:             slug,
			ShortDescription: shortDescription,
			Message:          message,
		})
		counts[secNum]++
	}
	if err := rows.Err(); err != nil {
		return model.SectionsResponse{}, err
	}
	if sections == nil {
		sections = []model.Section{}
	}
	return model.SectionsResponse{Sections: sections}, nil
}

func (s *ContentStore) SearchContent(ctx context.Context, q string) ([]model.Content, error) {
	like := "%" + q + "%"

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, execution_id, title, short_description, message, status,
		       category, sub_category, image_url, image_prompt, slug, created, last_updated
		FROM content
		WHERE status = 'PUBLISHED'
		  AND (title LIKE ? OR short_description LIKE ? OR slug LIKE ? OR message LIKE ?)
		ORDER BY
		  CASE
		    WHEN title             LIKE ? THEN 1
		    WHEN short_description LIKE ? THEN 2
		    WHEN slug              LIKE ? THEN 3
		    WHEN message           LIKE ? THEN 4
		    ELSE 5
		  END,
		  last_updated DESC
		LIMIT 20`,
		like, like, like, like, like, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.Content
	for rows.Next() {
		var c model.Content
		if err := rows.Scan(
			&c.ID, &c.ExecutionID, &c.Title, &c.ShortDescription, &c.Message,
			&c.Status, &c.Category, &c.SubCategory,
			&c.ImageURL, &c.ImagePrompt, &c.Slug, &c.Created, &c.LastUpdated,
		); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, rows.Err()
}

func (s *ContentStore) GetContentBySlug(ctx context.Context, slug string) (*model.Content, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, execution_id, title, short_description, message, status,
		       category, sub_category, image_url, image_prompt, slug, created, last_updated
		FROM content
		WHERE status = 'PUBLISHED' AND slug = ?
		LIMIT 1`, slug)

	var c model.Content
	err := row.Scan(
		&c.ID, &c.ExecutionID, &c.Title, &c.ShortDescription, &c.Message,
		&c.Status, &c.Category, &c.SubCategory,
		&c.ImageURL, &c.ImagePrompt, &c.Slug, &c.Created, &c.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
