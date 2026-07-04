package handler

import (
	"errors"
	"testing"

	"github.com/mattc/badger/internal/badge"
	"github.com/mattc/badger/internal/circleci"
)

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
	msg, color := badgeForError(circleci.ErrNoPipelines)
	if msg != "none" || color != badge.ColorLightGrey {
		t.Errorf("badgeForError(no pipelines) = (%q, %q), want (none, grey)", msg, color)
	}

	msg, color = badgeForError(errors.New("boom"))
	if msg != "error" || color != badge.ColorRed {
		t.Errorf("badgeForError(other) = (%q, %q), want (error, red)", msg, color)
	}
}
