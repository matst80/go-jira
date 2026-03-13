# Jira Agile API Client Example

This contains a reusable Go package for interacting with the Atlassian Jira REST API, alongside an example application that uses the package to query a specific agile board.

## Prerequisites

You need an Atlassian JWT scoped token provided via the `ATLASSIAN_KEY` environment variable.

## Packages

### `pkg/jira`

A reusable Jira API client that focuses on Atlassian's Agile API (e.g., fetching lists of issues for a board). To keep cyclomatic complexity low, it is organized into logical structs (`Client`, `ClientOption`) and concise data models (`SearchResult`, `Issue`, `Fields`).

Usage:

```go
client := jira.NewClient("https://your-domain.atlassian.net", "jwt-token")
results, err := client.ListBoardIssues(123, "status != Closed")
// Fetch a specific issue and its attachments
issue, err := client.GetIssue("PROJ-123")
// Download attachment bytes
data, err := client.DownloadAttachment(issue.Fields.Attachment[0])
```

```bash
# Provide the token
export ATLASSIAN_KEY="your_scoped_jwt_token_here"

# Run the app
go run ./cmd/example-jira/main.go
```
# go-jira
