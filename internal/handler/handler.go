package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/mattc/badger/internal/badge"
	"github.com/mattc/badger/internal/circleci"
)

// Handler serves badge endpoints.
type Handler struct {
	circleci *circleci.Client
}

// New creates a badge handler.
func New(circleciClient *circleci.Client) *Handler {
	return &Handler{circleci: circleciClient}
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

	branch := r.URL.Query().Get("branch")
	if branch == "" {
		writeBadge(w, "pipeline", "missing branch", badge.ColorRed)
		return
	}

	label := r.URL.Query().Get("label")
	if label == "" {
		label = "pipeline"
	}

	projectSlug := fmt.Sprintf("%s/%s/%s", vcs, org, repo)
	pipeline, err := h.circleci.LatestPipeline(r.Context(), projectSlug, branch)
	if err != nil {
		message, color := badgeForError(err)
		writeBadge(w, label, message, color)
		return
	}

	message := strconv.Itoa(pipeline.Number)
	color := pipelineColor(pipeline.State)
	writeBadge(w, label, message, color)
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
	if errors.Is(err, circleci.ErrNoPipelines) {
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
