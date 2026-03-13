package jira

// SearchResult represents the response from Jira agile board issue search
type SearchResult struct {
	Expand     string  `json:"expand,omitempty"`
	StartAt    int     `json:"startAt,omitempty"`
	MaxResults int     `json:"maxResults,omitempty"`
	Total      int     `json:"total,omitempty"`
	Issues     []Issue `json:"issues,omitempty"`
}

// Issue represents a single Jira issue
type Issue struct {
	Expand string `json:"expand,omitempty"`
	ID     string `json:"id"`
	Self   string `json:"self"`
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

// Fields contains the interesting properties of an issue
type Fields struct {
	Summary        string          `json:"summary"`
	Description    string          `json:"description,omitempty"`
	Status         Status          `json:"status"`
	IssueType      IssueType       `json:"issuetype"`
	Project        Project         `json:"project"`
	Priority       *Priority       `json:"priority,omitempty"`
	Assignee       *User           `json:"assignee,omitempty"`
	Reporter       *User           `json:"reporter,omitempty"`
	Creator        *User           `json:"creator,omitempty"`
	Created        string          `json:"created,omitempty"`
	Updated        string          `json:"updated,omitempty"`
	LastViewed     string          `json:"lastViewed,omitempty"`
	DueDate        string          `json:"duedate,omitempty"`
	Resolution     *Resolution     `json:"resolution,omitempty"`
	ResolutionDate string          `json:"resolutiondate,omitempty"`
	Labels         []string        `json:"labels,omitempty"`
	Components     []Component     `json:"components,omitempty"`
	Subtasks       []Issue         `json:"subtasks,omitempty"`
	Attachment     []Attachment    `json:"attachment,omitempty"`
	Progress       Progress        `json:"progress"`
	Worklog        Worklog         `json:"worklog"`
	Comment        Comment         `json:"comment"`
	Votes          Votes           `json:"votes"`
	Watches        Watches         `json:"watches"`
	TimeTracking   TimeTracking    `json:"timetracking"`
	Flagged        bool            `json:"flagged"`
	StatusCategory StatusCategory  `json:"statusCategory"`
	Sprint         interface{}     `json:"sprint,omitempty"`
	Epic           interface{}     `json:"epic,omitempty"`
	FixVersions    []Version       `json:"fixVersions,omitempty"`
	Versions       []Version       `json:"versions,omitempty"`
	IssueLinks     []interface{}   `json:"issuelinks,omitempty"`
	Environment    string          `json:"environment,omitempty"`
	Security       interface{}     `json:"security,omitempty"`

	// Custom Fields from JSON
	CustomField10190 *CustomFieldOption `json:"customfield_10190,omitempty"`
	CustomField10191 *CustomFieldOption `json:"customfield_10191,omitempty"`
	CustomField10199 *CustomFieldOption `json:"customfield_10199,omitempty"`
	CustomField10181 *CustomFieldOption `json:"customfield_10181,omitempty"`
	CustomField10176 *CustomFieldOption `json:"customfield_10176,omitempty"`
	CustomField10277 *CustomFieldOption `json:"customfield_10277,omitempty"`
	CustomField10216 *CustomFieldOption `json:"customfield_10216,omitempty"`
	CustomField10219 *CustomFieldOption `json:"customfield_10219,omitempty"`
	CustomField10211 *CustomFieldOption `json:"customfield_10211,omitempty"`
	CustomField10207 *CustomFieldOption `json:"customfield_10207,omitempty"`
	CustomField10145 *CustomFieldOption `json:"customfield_10145,omitempty"`
	CustomField12923 float64            `json:"customfield_12923,omitempty"`
	CustomField12922 string             `json:"customfield_12922,omitempty"`
	CustomField12806 string             `json:"customfield_12806,omitempty"`
	CustomField10829 *User              `json:"customfield_10829,omitempty"`
	CustomField10024 string             `json:"customfield_10024,omitempty"`
	CustomField10019 string             `json:"customfield_10019,omitempty"`
	CustomField10000 string             `json:"customfield_10000,omitempty"`
	CustomField10049 []interface{}      `json:"customfield_10049,omitempty"`
	CustomField10028 []interface{}      `json:"customfield_10028,omitempty"`
	CustomField10002 []interface{}      `json:"customfield_10002,omitempty"`
	CustomField10114 string             `json:"customfield_10114,omitempty"`
	CustomField10109 string             `json:"customfield_10109,omitempty"`
	CustomField10220 string             `json:"customfield_10220,omitempty"`
	CustomField10214 string             `json:"customfield_10214,omitempty"`
}

// User represents a Jira user
type User struct {
	Self         string      `json:"self"`
	AccountId    string      `json:"accountId"`
	EmailAddress string      `json:"emailAddress,omitempty"`
	AvatarUrls   *AvatarUrls `json:"avatarUrls,omitempty"`
	DisplayName  string      `json:"displayName"`
	Active       bool        `json:"active"`
	TimeZone     string      `json:"timeZone,omitempty"`
	AccountType  string      `json:"accountType,omitempty"`
}

// AvatarUrls contains links to user/project avatars
type AvatarUrls struct {
	X16 string `json:"16x16"`
	X24 string `json:"24x24"`
	X32 string `json:"32x32"`
	X48 string `json:"48x48"`
}

// Status represents the current state of the issue
type Status struct {
	Self           string         `json:"self"`
	Description    string         `json:"description,omitempty"`
	IconUrl        string         `json:"iconUrl,omitempty"`
	Name           string         `json:"name"`
	ID             string         `json:"id"`
	StatusCategory StatusCategory `json:"statusCategory"`
}

// StatusCategory represents the category of the status (To Do, In Progress, Done)
type StatusCategory struct {
	Self      string `json:"self"`
	ID        int    `json:"id"`
	Key       string `json:"key"`
	ColorName string `json:"colorName"`
	Name      string `json:"name"`
}

// IssueType represents the type of issue (Bug, Task, etc.)
type IssueType struct {
	Self           string `json:"self"`
	ID             string `json:"id"`
	Description    string `json:"description"`
	IconUrl        string `json:"iconUrl"`
	Name           string `json:"name"`
	Subtask        bool   `json:"subtask"`
	AvatarId       int    `json:"avatarId"`
	HierarchyLevel int    `json:"hierarchyLevel"`
}

// Project represents a Jira project
type Project struct {
	Self            string           `json:"self"`
	ID              string           `json:"id"`
	Key             string           `json:"key"`
	Name            string           `json:"name"`
	ProjectTypeKey  string           `json:"projectTypeKey"`
	Simplified      bool             `json:"simplified"`
	AvatarUrls      *AvatarUrls      `json:"avatarUrls"`
	ProjectCategory *ProjectCategory `json:"projectCategory,omitempty"`
}

// ProjectCategory represents a project's category
type ProjectCategory struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// Priority represents the priority of an issue
type Priority struct {
	Self    string `json:"self"`
	IconUrl string `json:"iconUrl"`
	Name    string `json:"name"`
	ID      string `json:"id"`
}

// Resolution represents the resolution of an issue
type Resolution struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// Component represents a Jira component
type Component struct {
	Self string `json:"self"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CustomFieldOption represents an option for custom fields
type CustomFieldOption struct {
	Self  string `json:"self"`
	Value string `json:"value"`
	ID    string `json:"id"`
}

// Progress represents the progress of an issue
type Progress struct {
	Progress int `json:"progress"`
	Total    int `json:"total"`
}

// Votes represents the votes on an issue
type Votes struct {
	Self     string `json:"self"`
	Votes    int    `json:"votes"`
	HasVoted bool   `json:"hasVoted"`
}

// Watches represents the watches on an issue
type Watches struct {
	Self       string `json:"self"`
	WatchCount int    `json:"watchCount"`
	IsWatching bool   `json:"isWatching"`
}

// Worklog represents the worklogs of an issue
type Worklog struct {
	StartAt    int              `json:"startAt"`
	MaxResults int              `json:"maxResults"`
	Total      int              `json:"total"`
	Worklogs   []WorklogContent `json:"worklogs"`
}

// WorklogContent represents a single worklog entry
type WorklogContent struct {
	Self             string `json:"self"`
	Author           User   `json:"author"`
	UpdateAuthor     User   `json:"updateAuthor"`
	Comment          string `json:"comment"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
	Started          string `json:"started"`
	TimeSpent        string `json:"timeSpent"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
	ID               string `json:"id"`
	IssueID          string `json:"issueId"`
}

// Comment represents the comments on an issue
type Comment struct {
	Self       string           `json:"self"`
	MaxResults int              `json:"maxResults"`
	Total      int              `json:"total"`
	StartAt    int              `json:"startAt"`
	Comments   []CommentContent `json:"comments"`
}

// CommentContent represents a single comment entry
type CommentContent struct {
	Self         string `json:"self"`
	ID           string `json:"id"`
	Author       User   `json:"author"`
	Body         string `json:"body"`
	UpdateAuthor User   `json:"updateAuthor"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
	JSDPublic    bool   `json:"jsdPublic"`
}

// Attachment represents a file attached to an issue
type Attachment struct {
	Self      string `json:"self"`
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Author    *User  `json:"author,omitempty"`
	Created   string `json:"created"`
	Size      int    `json:"size"`
	MimeType  string `json:"mimeType"`
	Content   string `json:"content"`
	Thumbnail string `json:"thumbnail"`
}

// TimeTracking represents timetracking information
type TimeTracking struct {
	OriginalEstimate         string `json:"originalEstimate,omitempty"`
	RemainingEstimate        string `json:"remainingEstimate,omitempty"`
	TimeSpent                string `json:"timeSpent,omitempty"`
	OriginalEstimateSeconds  int    `json:"originalEstimateSeconds,omitempty"`
	RemainingEstimateSeconds int    `json:"remainingEstimateSeconds,omitempty"`
	TimeSpentSeconds         int    `json:"timeSpentSeconds,omitempty"`
}

// Version represents a project version
type Version struct {
	Self        string `json:"self"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Archived    bool   `json:"archived"`
	Released    bool   `json:"released"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}
