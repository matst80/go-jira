package jira

import (
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
