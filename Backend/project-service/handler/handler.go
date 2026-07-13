package handler

import (
	"context"
	"errors"

	projectpb "forge/proto"
	"forge/project-service/model"
	"forge/project-service/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectHandler struct {
	projectpb.UnimplementedProjectServiceServer
	service *service.ProjectService
}

func NewProjectHandler(
	projectService *service.ProjectService,
) *ProjectHandler {
	return &ProjectHandler{
		service: projectService,
	}
}

func toProto(project *model.Project) *projectpb.ProjectResponse {
	return &projectpb.ProjectResponse{
		Id:        project.ID,
		Name:      project.Name,
		ImageName: project.ImageName,
		Status:    project.Status,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrProjectNotFound):
		return status.Error(codes.NotFound, "project not found")

	case errors.Is(err, service.ErrInvalidProject):
		return status.Error(
			codes.InvalidArgument,
			"name and image_name are required",
		)

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request timed out")

	default:
		return status.Error(codes.Internal, "internal service error")
	}
}

func (h *ProjectHandler) Health(
	ctx context.Context,
	req *projectpb.HealthRequest,
) (*projectpb.HealthResponse, error) {
	return &projectpb.HealthResponse{
		Status:  "ok",
		Service: "project-service",
	}, nil
}

func (h *ProjectHandler) GetProject(
	ctx context.Context,
	req *projectpb.GetProjectRequest,
) (*projectpb.ProjectResponse, error) {
	project, err := h.service.GetProject(ctx, req.GetId())
	if err != nil {
		return nil, mapError(err)
	}

	return toProto(project), nil
}

func (h *ProjectHandler) ListProjects(
	ctx context.Context,
	req *projectpb.ListProjectsRequest,
) (*projectpb.ListProjectsResponse, error) {
	projects, err := h.service.ListProjects(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	response := &projectpb.ListProjectsResponse{
		Projects: make([]*projectpb.ProjectResponse, 0, len(projects)),
	}

	for i := range projects {
		response.Projects = append(
			response.Projects,
			toProto(&projects[i]),
		)
	}

	return response, nil
}

func (h *ProjectHandler) CreateProject(
	ctx context.Context,
	req *projectpb.CreateProjectRequest,
) (*projectpb.ProjectResponse, error) {
	project, err := h.service.CreateProject(
		ctx,
		req.GetName(),
		req.GetImageName(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return toProto(project), nil
}

func (h *ProjectHandler) GetProjectURL(
	ctx context.Context,
	req *projectpb.GetProjectRequest,
) (*projectpb.ProjectURLResponse, error) {
	url, err := h.service.GetProjectURL(ctx, req.GetId())
	if err != nil {
		return nil, mapError(err)
	}

	return &projectpb.ProjectURLResponse{
		Url: url,
	}, nil
}