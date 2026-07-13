package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"forge/project-service/model"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetByID(
	ctx context.Context,
	id int32,
) (*model.Project, error) {
	project := &model.Project{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, image_name, status
		FROM projects
		WHERE id = $1
	`, id).Scan(
		&project.ID,
		&project.Name,
		&project.ImageName,
		&project.Status,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) List(
	ctx context.Context,
) ([]model.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, image_name, status
		FROM projects
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]model.Project, 0, 32)

	for rows.Next() {
		var project model.Project

		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.ImageName,
			&project.Status,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) Create(
	ctx context.Context,
	name string,
	imageName string,
) (*model.Project, error) {
	project := &model.Project{}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO projects (name, image_name, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, name, image_name, status
	`, name, imageName).Scan(
		&project.ID,
		&project.Name,
		&project.ImageName,
		&project.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}