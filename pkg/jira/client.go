package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
)

// Client for Atlassian Jira REST API
type Client struct {
	baseURL     string
	httpClient  *http.Client
	token       string
	email       string
	authPrefix  string
	debug       bool
	debugWriter io.Writer
}

// ClientOption allows customizing the client
type ClientOption func(*Client)

// WithBasicAuth allows using an API token with an email address
func WithBasicAuth(email string) ClientOption {
	return func(c *Client) {
		c.email = email
	}
}

// WithAuthPrefix allows using a custom prefix like "JWT" instead of "Bearer"
func WithAuthPrefix(prefix string) ClientOption {
	return func(c *Client) {
		c.authPrefix = prefix
	}
}

// WithHTTPClient allows using a custom http client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithDebug enables raw response logging to stdout
func WithDebug(enabled bool) ClientOption {
	return func(c *Client) {
		c.debug = enabled
	}
}

// WithDebugWriter enables raw response logging to a specific writer
func WithDebugWriter(w io.Writer) ClientOption {
	return func(c *Client) {
		c.debug = true
		c.debugWriter = w
	}
}

// NewClient initializes a new Jira API client.
// token should be a scoped JWT token for the Atlassian API.
// baseURL is the root domain (e.g., https://elkjop.atlassian.net).
func NewClient(baseURL, token string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ListBoardIssues fetches issues for a specific Jira Agile board.
// jql can be used to filter the issues (e.g., "status != Closed").
// Ref: https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-agile-1-0-board-boardid-issue-get
func (c *Client) ListBoardIssues(boardID int, jql string) (*SearchResult, error) {
	apiURL := fmt.Sprintf("%s/rest/agile/1.0/board/%d/issue", c.baseURL, boardID)
	return c.fetchIssues(apiURL, jql)
}

// ListBacklogIssues fetches issues in the backlog for a specific Jira Agile board.
// jql can be used to filter the issues (e.g., "description is EMPTY").
// Ref: https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/#api-agile-1-0-board-boardid-backlog-get
func (c *Client) ListBacklogIssues(boardID int, jql string) (*SearchResult, error) {
	apiURL := fmt.Sprintf("%s/rest/agile/1.0/board/%d/backlog", c.baseURL, boardID)
	return c.fetchIssues(apiURL, jql)
}

// GetIssue fetches a single Jira issue by ID or key.
func (c *Client) GetIssue(issueID string) (*Issue, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s", c.baseURL, issueID)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira api returned status %d", resp.StatusCode)
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return &issue, nil
}

// GetAttachmentMetadata fetches the metadata of a single attachment by ID.
func (c *Client) GetAttachmentMetadata(attachmentID string) (*Attachment, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/attachment/%s", c.baseURL, attachmentID)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira api returned status %d", resp.StatusCode)
	}

	var attachment Attachment
	if err := json.NewDecoder(resp.Body).Decode(&attachment); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return &attachment, nil
}

// DownloadAttachment downloads the content of an attachment.
func (c *Client) DownloadAttachment(attachment Attachment) ([]byte, error) {
	if attachment.Content == "" {
		return nil, fmt.Errorf("attachment has no content URL")
	}

	req, err := http.NewRequest(http.MethodGet, attachment.Content, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira api returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// SaveImage downloads and saves an attachment to the local filesystem.
func (c *Client) SaveImage(attachment Attachment) error {
	content, err := c.DownloadAttachment(attachment)
	if err != nil {
		return err
	}
	return os.WriteFile(attachment.Filename, content, 0644)
}

// AddComment adds a comment to a Jira issue.
func (c *Client) AddComment(issueID string, comment string) error {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s/comment", c.baseURL, issueID)

	payload := struct {
		Body string `json:"body"`
	}{
		Body: comment,
	}

	return c.doJSONRequest(http.MethodPost, apiURL, payload)
}

// AssignIssue assigns a Jira issue to a user by account ID.
func (c *Client) AssignIssue(issueID string, assignee string) error {
	return c.AssignIssueByAccountID(issueID, assignee)
}

// AssignIssueByAccountID assigns a Jira issue using a Jira Cloud account ID.
func (c *Client) AssignIssueByAccountID(issueID string, accountID string) error {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s/assignee", c.baseURL, issueID)

	payload := struct {
		AccountID string `json:"accountId,omitempty"`
	}{
		AccountID: accountID,
	}

	return c.doJSONRequest(http.MethodPut, apiURL, payload)
}

// AssignIssueByName assigns a Jira issue using a legacy Jira username.
func (c *Client) AssignIssueByName(issueID string, username string) error {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s/assignee", c.baseURL, issueID)

	payload := struct {
		Name string `json:"name,omitempty"`
	}{
		Name: username,
	}

	return c.doJSONRequest(http.MethodPut, apiURL, payload)
}

// FindUsers searches Jira users and returns matching users with account IDs.
func (c *Client) FindUsers(query string) ([]User, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/user/search?query=%s", c.baseURL, url.QueryEscape(query))
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return nil, fmt.Errorf("jira api returned status %d: %v", resp.StatusCode, body)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return users, nil
}

// GetTransitions fetches the available transitions for a Jira issue.
func (c *Client) GetTransitions(issueID string) (*TransitionsResult, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s/transitions", c.baseURL, issueID)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return nil, fmt.Errorf("jira api returned status %d: %v", resp.StatusCode, body)
	}

	var result TransitionsResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response failed: %w", err)
	}

	return &result, nil
}

// TransitionIssue transitions a Jira issue using a transition ID.
func (c *Client) TransitionIssue(issueID string, transitionID string) error {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s/transitions", c.baseURL, issueID)

	payload := struct {
		Transition struct {
			ID string `json:"id"`
		} `json:"transition"`
	}{}
	payload.Transition.ID = transitionID

	return c.doJSONRequest(http.MethodPost, apiURL, payload)
}

func (c *Client) setAuth(req *http.Request) {
	if c.email != "" {
		req.SetBasicAuth(c.email, c.token)
	} else {
		prefix := "Bearer"
		if c.authPrefix != "" {
			prefix = c.authPrefix
		}
		req.Header.Set("Authorization", prefix+" "+c.token)
	}
}

func (c *Client) fetchIssues(apiURL, jql string) (*SearchResult, error) {
	if jql != "" {
		apiURL = fmt.Sprintf("%s?jql=%s", apiURL, url.QueryEscape(jql))
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return nil, fmt.Errorf("jira api returned status %d: %v", resp.StatusCode, body)
	}

	var bodyBytes []byte
	if c.debug {
		var err error
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body failed: %w", err)
		}
		if c.debugWriter != nil {
			fmt.Fprintf(c.debugWriter, "DEBUG: Raw Response: %s\n", string(bodyBytes))
		} else {
			fmt.Printf("DEBUG: Raw Response: %s\n", string(bodyBytes))
		}
	}

	var result SearchResult
	if c.debug {
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("decoding response failed: %w", err)
		}
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decoding response failed: %w", err)
		}
	}

	return &result, nil
}

func (c *Client) doJSONRequest(method, apiURL string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding request failed: %w", err)
	}

	req, err := http.NewRequest(method, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	c.setAuth(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var responseBody interface{}
		_ = json.NewDecoder(resp.Body).Decode(&responseBody)
		return fmt.Errorf("jira api returned status %d: %v", resp.StatusCode, responseBody)
	}

	return nil
}
