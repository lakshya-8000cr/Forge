package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"forge/project-service/model"
	"forge/project-service/repository"
)

var (
	ErrInvalidProject   = errors.New("name and image_name are required")
	ErrProjectNotFound  = repository.ErrProjectNotFound
)

type ProjectService struct {
	repository *repository.ProjectRepository
	publicURL  string
}

func NewProjectService(
	repository *repository.ProjectRepository,
	publicURL string,
) *ProjectService {
	return &ProjectService{
		repository: repository,
		publicURL:  strings.TrimRight(publicURL, "/"),
	}
}

func (s *ProjectService) GetProject(
	ctx context.Context,
	id int32,
) (*model.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repository.GetByID(ctx, id)
}

func (s *ProjectService) ListProjects(
	ctx context.Context,
) ([]model.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repository.List(ctx)
}

func (s *ProjectService) CreateProject(
	ctx context.Context,
	name string,
	imageName string,
) (*model.Project, error) {
	if strings.TrimSpace(name) == "" ||
		strings.TrimSpace(imageName) == "" {
		return nil, ErrInvalidProject
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repository.Create(ctx, name, imageName)
}

func (s *ProjectService) GetProjectURL(
	ctx context.Context,
	id int32,
) (string, error) {
	project, err := s.GetProject(ctx, id)
	if err != nil {
		return "", err
	}

	return s.publicURL + "/apps/" + project.Name, nil
}