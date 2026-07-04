package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattsre/badger/internal/badge"
	"github.com/mattsre/badger/internal/circleci"
)

type stubPipelineClient struct {
	pipeline *circleci.Pipeline
	err      error
}

func (s *stubPipelineClient) LatestPipeline(context.Context, string, string) (*circleci.Pipeline, error) {
	return s.pipeline, s.err
}

func TestFormatPipelineValue(t *testing.T) {
	tests := []struct {
		template string
		number   int
		want     string
	}{
		{"", 42, "42"},
		{"0.1.$PIPELINE_NUMBER", 42, "0.1.42"},
		{"v{number}-rc", 7, "v7-rc"},
		{"$PIPELINE_NUMBER", 99, "99"},
	}

	for _, tt := range tests {
		if got := formatPipelineValue(tt.template, tt.number); got != tt.want {
			t.Errorf("formatPipelineValue(%q, %d) = %q, want %q", tt.template, tt.number, got, tt.want)
		}
	}
}

func TestCircleciPipelineBadge(t *testing.T) {
	h := New(&stubPipelineClient{
		pipeline: &circleci.Pipeline{Number: 42, State: "success"},
	})

	req := httptest.NewRequest(http.MethodGet, "/circleci/gh/myorg/myrepo/pipeline?branch=main&label=tag&value=0.1.$PIPELINE_NUMBER", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, ">tag<") {
		t.Errorf("expected custom label in SVG, got: %s", body)
	}
	if !strings.Contains(body, ">0.1.42<") {
		t.Errorf("expected formatted value in SVG, got: %s", body)
	}
}

func TestCircleciPipelineBadgeMessageAlias(t *testing.T) {
	h := New(&stubPipelineClient{
		pipeline: &circleci.Pipeline{Number: 5, State: "success"},
	})

	req := httptest.NewRequest(http.MethodGet, "/circleci/gh/myorg/myrepo/pipeline?branch=main&label=tag&message=0.1.$PIPELINE_NUMBER", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, ">0.1.5<") {
		t.Errorf("expected message query param to format value, got: %s", body)
	}
}

func TestCircleciPipelineMissingBranchUsesLabel(t *testing.T) {
	h := New(&stubPipelineClient{})

	req := httptest.NewRequest(http.MethodGet, "/circleci/gh/myorg/myrepo/pipeline?label=tag", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, ">tag<") {
		t.Errorf("expected custom label for missing branch, got: %s", body)
	}
	if !strings.Contains(body, ">missing branch<") {
		t.Errorf("expected missing branch message, got: %s", body)
	}
}

func TestParseCircleCIProjectPath(t *testing.T) {
	tests := []struct {
		path     string
		wantVCS  string
		wantOrg  string
		wantRepo string
		wantOK   bool
	}{
		{
			path:     "/circleci/gh/myorg/myrepo/pipeline",
			wantVCS:  "gh",
			wantOrg:  "myorg",
			wantRepo: "myrepo",
			wantOK:   true,
		},
		{
			path:   "/circleci/gh/org/repo/wrong",
			wantOK: false,
		},
		{
			path:   "/circleci/gh/org/repo/pipeline/extra",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		vcs, org, repo, ok := parseCircleCIProjectPath(tt.path)
		if ok != tt.wantOK {
			t.Errorf("parseCircleCIProjectPath(%q) ok = %v, want %v", tt.path, ok, tt.wantOK)
			continue
		}
		if !tt.wantOK {
			continue
		}
		if vcs != tt.wantVCS || org != tt.wantOrg || repo != tt.wantRepo {
			t.Errorf("parseCircleCIProjectPath(%q) = (%q,%q,%q), want (%q,%q,%q)",
				tt.path, vcs, org, repo, tt.wantVCS, tt.wantOrg, tt.wantRepo)
		}
	}
}

func TestPipelineColor(t *testing.T) {
	if pipelineColor("success") != "#4c1" {
		t.Error("expected green for success")
	}
	if pipelineColor("failed") != "#e05d44" {
		t.Error("expected red for failed")
	}
}

func TestBadgeForError(t *testing.T) {
	msg, color := badgeForError(&circleci.NoPipelinesError{Branch: "main"})
	if msg != "none" || color != badge.ColorLightGrey {
		t.Errorf("badgeForError(no pipelines) = (%q, %q), want (none, grey)", msg, color)
	}

	msg, color = badgeForError(errors.New("boom"))
	if msg != "error" || color != badge.ColorRed {
		t.Errorf("badgeForError(other) = (%q, %q), want (error, red)", msg, color)
	}
}
