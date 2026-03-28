package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matst80/go-jira/pkg/jira"
)

type transitionLister interface {
	GetTransitions(issueID string) (*jira.TransitionsResult, error)
}

func TestExampleTagQueryFlow(t *testing.T) {
	var requestedPath string
	var requestedJQL string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedJQL = r.URL.Query().Get("jql")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"startAt": 0,
			"maxResults": 50,
			"total": 2,
			"issues": [
				{
					"id": "10001",
					"key": "AI-1",
					"self": "http://example.test/rest/api/2/issue/10001",
					"fields": {
						"summary": "First issue",
						"description": "desc",
						"status": {
							"id": "1",
							"name": "To Do",
							"self": "",
							"statusCategory": {
								"id": 2,
								"key": "new",
								"colorName": "blue-gray",
								"name": "To Do",
								"self": ""
							}
						},
						"issuetype": {
							"id": "10",
							"name": "Task",
							"description": "",
							"iconUrl": "",
							"self": "",
							"subtask": false,
							"avatarId": 0,
							"hierarchyLevel": 0
						},
						"project": {
							"id": "200",
							"key": "AI",
							"name": "AI",
							"self": "",
							"projectTypeKey": "software",
							"simplified": false,
							"avatarUrls": null
						},
						"created": "2024-01-01T00:00:00.000+0000",
						"updated": "2024-01-02T00:00:00.000+0000",
						"comment": {
							"self": "",
							"maxResults": 0,
							"total": 0,
							"startAt": 0,
							"comments": []
						},
						"attachment": [],
						"progress": {
							"progress": 0,
							"total": 0
						},
						"worklog": {
							"startAt": 0,
							"maxResults": 0,
							"total": 0,
							"worklogs": []
						},
						"votes": {
							"self": "",
							"votes": 0,
							"hasVoted": false
						},
						"watches": {
							"self": "",
							"watchCount": 0,
							"isWatching": false
						},
						"timetracking": {},
						"flagged": false,
						"statusCategory": {
							"id": 2,
							"key": "new",
							"colorName": "blue-gray",
							"name": "To Do",
							"self": ""
						}
					}
				},
				{
					"id": "10002",
					"key": "AI-2",
					"self": "http://example.test/rest/api/2/issue/10002",
					"fields": {
						"summary": "Second issue",
						"description": "desc 2",
						"status": {
							"id": "3",
							"name": "In Progress",
							"self": "",
							"statusCategory": {
								"id": 4,
								"key": "indeterminate",
								"colorName": "yellow",
								"name": "In Progress",
								"self": ""
							}
						},
						"issuetype": {
							"id": "10",
							"name": "Task",
							"description": "",
							"iconUrl": "",
							"self": "",
							"subtask": false,
							"avatarId": 0,
							"hierarchyLevel": 0
						},
						"project": {
							"id": "200",
							"key": "AI",
							"name": "AI",
							"self": "",
							"projectTypeKey": "software",
							"simplified": false,
							"avatarUrls": null
						},
						"created": "2024-01-03T00:00:00.000+0000",
						"updated": "2024-01-04T00:00:00.000+0000",
						"comment": {
							"self": "",
							"maxResults": 0,
							"total": 0,
							"startAt": 0,
							"comments": []
						},
						"attachment": [],
						"progress": {
							"progress": 0,
							"total": 0
						},
						"worklog": {
							"startAt": 0,
							"maxResults": 0,
							"total": 0,
							"worklogs": []
						},
						"votes": {
							"self": "",
							"votes": 0,
							"hasVoted": false
						},
						"watches": {
							"self": "",
							"watchCount": 0,
							"isWatching": false
						},
						"timetracking": {},
						"flagged": false,
						"statusCategory": {
							"id": 4,
							"key": "indeterminate",
							"colorName": "yellow",
							"name": "In Progress",
							"self": ""
						}
					}
				}
			]
		}`))
	}))
	defer ts.Close()

	client := jira.NewClient(ts.URL, "dummy-token")
	jql := `labels = ai-task AND statusCategory != Done ORDER BY created DESC`

	issues, err := client.ListBoardIssues(773, jql)
	if err != nil {
		t.Fatalf("ListBoardIssues failed: %v", err)
	}

	if requestedPath != "/rest/agile/1.0/board/773/issue" {
		t.Fatalf("expected path %q, got %q", "/rest/agile/1.0/board/773/issue", requestedPath)
	}

	if requestedJQL != jql {
		t.Fatalf("expected JQL %q, got %q", jql, requestedJQL)
	}

	if issues.Total != 2 {
		t.Fatalf("expected total 2, got %d", issues.Total)
	}

	if len(issues.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues.Issues))
	}

	if issues.Issues[1].Key != "AI-2" {
		t.Fatalf("expected second issue %q, got %q", "AI-2", issues.Issues[1].Key)
	}
}

func TestExampleIssueActionsAgainstFetchedIssue(t *testing.T) {
	var commentCalled bool
	var assignCalled bool
	var transitionCalled bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/agile/1.0/board/773/issue":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"startAt": 0,
				"maxResults": 50,
				"total": 2,
				"issues": [
					{
						"id": "10001",
						"key": "AI-1",
						"self": "",
						"fields": {
							"summary": "First issue",
							"description": "desc",
							"status": {
								"id": "1",
								"name": "To Do",
								"self": "",
								"statusCategory": {
									"id": 2,
									"key": "new",
									"colorName": "blue-gray",
									"name": "To Do",
									"self": ""
								}
							},
							"issuetype": {
								"id": "10",
								"name": "Task",
								"description": "",
								"iconUrl": "",
								"self": "",
								"subtask": false,
								"avatarId": 0,
								"hierarchyLevel": 0
							},
							"project": {
								"id": "200",
								"key": "AI",
								"name": "AI",
								"self": "",
								"projectTypeKey": "software",
								"simplified": false,
								"avatarUrls": null
							},
							"created": "2024-01-01T00:00:00.000+0000",
							"updated": "2024-01-02T00:00:00.000+0000",
							"comment": {
								"self": "",
								"maxResults": 0,
								"total": 0,
								"startAt": 0,
								"comments": []
							},
							"attachment": [],
							"progress": {
								"progress": 0,
								"total": 0
							},
							"worklog": {
								"startAt": 0,
								"maxResults": 0,
								"total": 0,
								"worklogs": []
							},
							"votes": {
								"self": "",
								"votes": 0,
								"hasVoted": false
							},
							"watches": {
								"self": "",
								"watchCount": 0,
								"isWatching": false
							},
							"timetracking": {},
							"flagged": false,
							"statusCategory": {
								"id": 2,
								"key": "new",
								"colorName": "blue-gray",
								"name": "To Do",
								"self": ""
							}
						}
					},
					{
						"id": "10002",
						"key": "AI-2",
						"self": "",
						"fields": {
							"summary": "Second issue",
							"description": "desc 2",
							"status": {
								"id": "3",
								"name": "In Progress",
								"self": "",
								"statusCategory": {
									"id": 4,
									"key": "indeterminate",
									"colorName": "yellow",
									"name": "In Progress",
									"self": ""
								}
							},
							"issuetype": {
								"id": "10",
								"name": "Task",
								"description": "",
								"iconUrl": "",
								"self": "",
								"subtask": false,
								"avatarId": 0,
								"hierarchyLevel": 0
							},
							"project": {
								"id": "200",
								"key": "AI",
								"name": "AI",
								"self": "",
								"projectTypeKey": "software",
								"simplified": false,
								"avatarUrls": null
							},
							"created": "2024-01-03T00:00:00.000+0000",
							"updated": "2024-01-04T00:00:00.000+0000",
							"comment": {
								"self": "",
								"maxResults": 0,
								"total": 0,
								"startAt": 0,
								"comments": []
							},
							"attachment": [],
							"progress": {
								"progress": 0,
								"total": 0
							},
							"worklog": {
								"startAt": 0,
								"maxResults": 0,
								"total": 0,
								"worklogs": []
							},
							"votes": {
								"self": "",
								"votes": 0,
								"hasVoted": false
							},
							"watches": {
								"self": "",
								"watchCount": 0,
								"isWatching": false
							},
							"timetracking": {},
							"flagged": false,
							"statusCategory": {
								"id": 4,
								"key": "indeterminate",
								"colorName": "yellow",
								"name": "In Progress",
								"self": ""
							}
						}
					}
				]
			}`))
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/2/issue/AI-2/comment":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading comment body failed: %v", err)
			}
			if got, want := strings.TrimSpace(string(body)), `{"body":"test comment"}`; got != want {
				t.Fatalf("unexpected comment body: got %q want %q", got, want)
			}
			commentCalled = true
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPut && r.URL.Path == "/rest/api/2/issue/AI-2/assignee":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading assign body failed: %v", err)
			}
			if got, want := strings.TrimSpace(string(body)), `{"accountId":"alice"}`; got != want {
				t.Fatalf("unexpected assign body: got %q want %q", got, want)
			}
			assignCalled = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/2/issue/AI-2/transitions":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading transition body failed: %v", err)
			}
			if got, want := strings.TrimSpace(string(body)), `{"transition":{"id":"31"}}`; got != want {
				t.Fatalf("unexpected transition body: got %q want %q", got, want)
			}
			transitionCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	client := jira.NewClient(ts.URL, "dummy-token")
	issues, err := client.ListBoardIssues(773, `labels = ai-task AND statusCategory != Done ORDER BY created DESC`)
	if err != nil {
		t.Fatalf("ListBoardIssues failed: %v", err)
	}

	if len(issues.Issues) <= 1 {
		t.Fatalf("expected at least 2 issues, got %d", len(issues.Issues))
	}

	target := issues.Issues[1].Key
	if target != "AI-2" {
		t.Fatalf("expected target issue %q, got %q", "AI-2", target)
	}

	if err := client.AddComment(target, "test comment"); err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
	if err := client.AssignIssue(target, "alice"); err != nil {
		t.Fatalf("AssignIssue failed: %v", err)
	}
	if err := client.TransitionIssue(target, "31"); err != nil {
		t.Fatalf("TransitionIssue failed: %v", err)
	}

	if !commentCalled {
		t.Fatal("expected AddComment request to be made")
	}
	if !assignCalled {
		t.Fatal("expected AssignIssue request to be made")
	}
	if !transitionCalled {
		t.Fatal("expected TransitionIssue request to be made")
	}
}

func TestExampleTransitionsListing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected method %q, got %q", http.MethodGet, r.Method)
		}
		expectedPath := "/rest/api/2/issue/AI-2/transitions"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"transitions": [
				{
					"id": "11",
					"name": "Start Progress",
					"to": {
						"id": "3",
						"name": "In Progress",
						"self": "",
						"statusCategory": {
							"id": 4,
							"key": "indeterminate",
							"colorName": "yellow",
							"name": "In Progress",
							"self": ""
						}
					}
				},
				{
					"id": "31",
					"name": "Done",
					"to": {
						"id": "10000",
						"name": "Done",
						"self": "",
						"statusCategory": {
							"id": 3,
							"key": "done",
							"colorName": "green",
							"name": "Done",
							"self": ""
						}
					}
				}
			]
		}`))
	}))
	defer ts.Close()

	client := jira.NewClient(ts.URL, "dummy-token")

	lister, ok := interface{}(client).(transitionLister)
	if !ok {
		t.Fatal("client does not implement GetTransitions(issueID string)")
	}

	result, err := lister.GetTransitions("AI-2")
	if err != nil {
		t.Fatalf("GetTransitions failed: %v", err)
	}

	if len(result.Transitions) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(result.Transitions))
	}

	if result.Transitions[0].ID != "11" {
		t.Fatalf("expected first transition ID %q, got %q", "11", result.Transitions[0].ID)
	}

	if result.Transitions[0].Name != "Start Progress" {
		t.Fatalf("expected first transition name %q, got %q", "Start Progress", result.Transitions[0].Name)
	}

	if result.Transitions[1].ID != "31" {
		t.Fatalf("expected second transition ID %q, got %q", "31", result.Transitions[1].ID)
	}

	if result.Transitions[1].To.Name != "Done" {
		t.Fatalf("expected destination status %q, got %q", "Done", result.Transitions[1].To.Name)
	}
}

func TestSaveImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	content := []byte("fake-image-data")

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	SaveImage(content, "image.png")

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(got) != string(content) {
		t.Fatalf("unexpected file content: got %q want %q", string(got), string(content))
	}
}
