package circleci

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://circleci.com/api/v2"

// NoPipelinesError is returned when a branch has no CircleCI pipelines.
type NoPipelinesError struct {
	Branch string
}

func (e *NoPipelinesError) Error() string {
	return fmt.Sprintf("no pipelines found for branch %q", e.Branch)
}

// Client talks to the CircleCI API v2.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a CircleCI API client. token may be empty for public projects.
func NewClient(token string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Pipeline holds the fields we care about from a CircleCI pipeline.
type Pipeline struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	State  string `json:"state"`
}

type pipelineListResponse struct {
	Items []Pipeline `json:"items"`
}

// LatestPipeline returns the most recent pipeline for projectSlug (e.g. "gh/org/repo")
// on the given branch.
func (c *Client) LatestPipeline(ctx context.Context, projectSlug, branch string) (*Pipeline, error) {
	endpoint := fmt.Sprintf("%s/project/%s/pipeline", c.baseURL, escapeSlug(projectSlug))
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("branch", branch)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Circle-Token", c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("circleci API returned %d: %s", resp.StatusCode, string(body))
	}

	var list pipelineListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, &NoPipelinesError{Branch: branch}
	}

	return &list.Items[0], nil
}

func escapeSlug(slug string) string {
	parts := strings.Split(slug, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
