package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListBoardIssues_JQL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedJQL := "status != Closed"
		query := r.URL.Query().Get("jql")
		if query != expectedJQL {
			t.Errorf("Expected JQL %q, got %q", expectedJQL, query)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total": 1, "issues": []}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	_, err := client.ListBoardIssues(1, "status != Closed")
	if err != nil {
		t.Fatalf("ListBoardIssues failed: %v", err)
	}
}

func TestListBacklogIssues(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/rest/agile/1.0/board/456/backlog"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}
		expectedJQL := "description is EMPTY"
		query := r.URL.Query().Get("jql")
		if query != expectedJQL {
			t.Errorf("Expected JQL %q, got %q", expectedJQL, query)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total": 5, "issues": []}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	res, err := client.ListBacklogIssues(456, "description is EMPTY")
	if err != nil {
		t.Fatalf("ListBacklogIssues failed: %v", err)
	}
	if res.Total != 5 {
		t.Errorf("Expected total 5, got %d", res.Total)
	}
}

func TestAddComment(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method %q, got %q", http.MethodPost, r.Method)
		}
		expectedPath := "/rest/api/2/issue/ISSUE-123/comment"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer dummy-token" {
			t.Fatalf("Expected authorization header %q, got %q", "Bearer dummy-token", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Expected content type %q, got %q", "application/json", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Reading request body failed: %v", err)
		}
		expectedBody := `{"body":"hello from test"}`
		if string(body) != expectedBody {
			t.Fatalf("Expected body %q, got %q", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	if err := client.AddComment("ISSUE-123", "hello from test"); err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
}

func TestAssignIssue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("Expected method %q, got %q", http.MethodPut, r.Method)
		}
		expectedPath := "/rest/api/2/issue/ISSUE-456/assignee"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Reading request body failed: %v", err)
		}
		expectedBody := `{"accountId":"alice"}`
		if string(body) != expectedBody {
			t.Fatalf("Expected body %q, got %q", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	if err := client.AssignIssue("ISSUE-456", "alice"); err != nil {
		t.Fatalf("AssignIssue failed: %v", err)
	}
}

func TestTransitionIssue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected method %q, got %q", http.MethodPost, r.Method)
		}
		expectedPath := "/rest/api/2/issue/ISSUE-789/transitions"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Reading request body failed: %v", err)
		}
		expectedBody := `{"transition":{"id":"31"}}`
		if string(body) != expectedBody {
			t.Fatalf("Expected body %q, got %q", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	if err := client.TransitionIssue("ISSUE-789", "31"); err != nil {
		t.Fatalf("TransitionIssue failed: %v", err)
	}
}

func TestGetTransitions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected method %q, got %q", http.MethodGet, r.Method)
		}
		expectedPath := "/rest/api/2/issue/ISSUE-999/transitions"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer dummy-token" {
			t.Fatalf("Expected authorization header %q, got %q", "Bearer dummy-token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("Expected accept header %q, got %q", "application/json", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"transitions": [
				{
					"id": "11",
					"name": "Start Progress",
					"to": {
						"id": "3",
						"name": "In Progress"
					}
				},
				{
					"id": "21",
					"name": "Done",
					"to": {
						"id": "10000",
						"name": "Done"
					}
				}
			]
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	transitions, err := client.GetTransitions("ISSUE-999")
	if err != nil {
		t.Fatalf("GetTransitions failed: %v", err)
	}

	if len(transitions.Transitions) != 2 {
		t.Fatalf("Expected 2 transitions, got %d", len(transitions.Transitions))
	}

	if transitions.Transitions[0].ID != "11" {
		t.Fatalf("Expected first transition ID %q, got %q", "11", transitions.Transitions[0].ID)
	}

	if transitions.Transitions[0].Name != "Start Progress" {
		t.Fatalf("Expected first transition name %q, got %q", "Start Progress", transitions.Transitions[0].Name)
	}

	if transitions.Transitions[0].To.ID != "3" {
		t.Fatalf("Expected first transition target status ID %q, got %q", "3", transitions.Transitions[0].To.ID)
	}

	if transitions.Transitions[0].To.Name != "In Progress" {
		t.Fatalf("Expected first transition target status name %q, got %q", "In Progress", transitions.Transitions[0].To.Name)
	}
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected method %q, got %q", http.MethodGet, r.Method)
		}
		expectedPath := "/rest/api/3/user/search"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}
		expectedQuery := "mats.tornberg@elkjop.no"
		if got := r.URL.Query().Get("query"); got != expectedQuery {
			t.Fatalf("Expected query %q, got %q", expectedQuery, got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer dummy-token" {
			t.Fatalf("Expected authorization header %q, got %q", "Bearer dummy-token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("Expected accept header %q, got %q", "application/json", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"self": "https://example.atlassian.net/rest/api/3/user?accountId=abc123",
				"accountId": "abc123",
				"emailAddress": "mats.tornberg@elkjop.no",
				"displayName": "Mats Tornberg",
				"active": true,
				"accountType": "atlassian"
			}
		]`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	users, err := client.FindUsers("mats.tornberg@elkjop.no")
	if err != nil {
		t.Fatalf("FindUsers failed: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	if users[0].AccountId != "abc123" {
		t.Fatalf("Expected account ID %q, got %q", "abc123", users[0].AccountId)
	}

	if users[0].DisplayName != "Mats Tornberg" {
		t.Fatalf("Expected display name %q, got %q", "Mats Tornberg", users[0].DisplayName)
	}

	if users[0].EmailAddress != "mats.tornberg@elkjop.no" {
		t.Fatalf("Expected email %q, got %q", "mats.tornberg@elkjop.no", users[0].EmailAddress)
	}
}

func TestAssignIssueByAccountID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("Expected method %q, got %q", http.MethodPut, r.Method)
		}
		expectedPath := "/rest/api/2/issue/ISSUE-456/assignee"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Reading request body failed: %v", err)
		}
		expectedBody := `{"accountId":"abc123"}`
		if string(body) != expectedBody {
			t.Fatalf("Expected body %q, got %q", expectedBody, string(body))
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dummy-token")
	if err := client.AssignIssueByAccountID("ISSUE-456", "abc123"); err != nil {
		t.Fatalf("AssignIssueByAccountID failed: %v", err)
	}
}
