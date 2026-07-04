package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/mattsre/badger/internal/badge"
	"github.com/mattsre/badger/internal/circleci"
)

// pipelineFetcher loads the latest CircleCI pipeline for a project branch.
type pipelineFetcher interface {
	LatestPipeline(ctx context.Context, projectSlug, branch string) (*circleci.Pipeline, error)
}

// Handler serves badge endpoints.
type Handler struct {
	circleci        pipelineFetcher
	allowedProjects map[string]struct{}
	allowedBranches map[string]struct{}
}

// New creates a badge handler.
func New(circleciClient pipelineFetcher, allowedProjects, allowedBranches []string) *Handler {
	projects := make(map[string]struct{}, len(allowedProjects))
	for _, project := range allowedProjects {
		projects[project] = struct{}{}
	}
	branches := make(map[string]struct{}, len(allowedBranches))
	for _, branch := range allowedBranches {
		branches[branch] = struct{}{}
	}

	return &Handler{
		circleci:        circleciClient,
		allowedProjects: projects,
		allowedBranches: branches,
	}
}

// ServeHTTP routes badge requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/healthz":
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	case strings.HasPrefix(r.URL.Path, "/circleci/"):
		h.circleciPipeline(w, r)
	default:
		http.NotFound(w, r)
	}
}

// circleciPipeline handles:
//
//	/circleci/{vcs}/{org}/{repo}/pipeline?branch={branch}
//
// Example:
//
//	/circleci/gh/myorg/myrepo/pipeline?branch=main
func (h *Handler) circleciPipeline(w http.ResponseWriter, r *http.Request) {
	vcs, org, repo, ok := parseCircleCIProjectPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	projectSlug := fmt.Sprintf("%s/%s/%s", vcs, org, repo)
	if !h.projectAllowed(projectSlug) {
		http.NotFound(w, r)
		return
	}

	q := r.URL.Query()
	label := q.Get("label")
	if label == "" {
		label = "pipeline"
	}

	branch := q.Get("branch")
	if branch == "" {
		writeBadge(w, label, "missing branch", badge.ColorRed)
		return
	}
	if !h.branchAllowed(branch) {
		http.NotFound(w, r)
		return
	}

	valueTemplate := q.Get("value")
	if valueTemplate == "" {
		valueTemplate = q.Get("message")
	}

	pipeline, err := h.circleci.LatestPipeline(r.Context(), projectSlug, branch)
	if err != nil {
		message, color := badgeForError(err)
		writeBadge(w, label, message, color)
		return
	}

	message := formatPipelineValue(valueTemplate, pipeline.Number)
	color := pipelineColor(pipeline.State)
	writeBadge(w, label, message, color)
}

func (h *Handler) projectAllowed(projectSlug string) bool {
	_, ok := h.allowedProjects[projectSlug]
	return ok
}

func (h *Handler) branchAllowed(branch string) bool {
	_, ok := h.allowedBranches[branch]
	return ok
}

// formatPipelineValue renders the badge message from an optional template.
// Use $PIPELINE_NUMBER or {number} as placeholders for the pipeline number.
func formatPipelineValue(template string, number int) string {
	if template == "" {
		return strconv.Itoa(number)
	}
	num := strconv.Itoa(number)
	s := strings.ReplaceAll(template, "$PIPELINE_NUMBER", num)
	s = strings.ReplaceAll(s, "{number}", num)
	return s
}

func parseCircleCIProjectPath(path string) (vcs, org, repo string, ok bool) {
	// /circleci/{vcs}/{org}/{repo}/pipeline
	const prefix = "/circleci/"
	if !strings.HasPrefix(path, prefix) {
		return
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 4 || parts[3] != "pipeline" {
		return
	}
	return parts[0], parts[1], parts[2], true
}

func badgeForError(err error) (message, color string) {
	if _, ok := errors.AsType[*circleci.NoPipelinesError](err); ok {
		return "none", badge.ColorLightGrey
	}
	return "error", badge.ColorRed
}

func pipelineColor(state string) string {
	switch strings.ToLower(state) {
	case "success", "created":
		return badge.ColorBrightGreen
	case "running", "pending", "setup":
		return badge.ColorYellow
	case "failed", "error", "failing":
		return badge.ColorRed
	case "canceled", "cancelled":
		return badge.ColorLightGrey
	default:
		return badge.ColorBlue
	}
}

func writeBadge(w http.ResponseWriter, label, message, color string) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(badge.SVG(label, message, color)))
}
