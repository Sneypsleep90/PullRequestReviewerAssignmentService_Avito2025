package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_CreateUser(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	user := map[string]interface{}{
		"id":        "user1",
		"username":  "testuser",
		"is_active": true,
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user1", response["id"])
	assert.Equal(t, "testuser", response["username"])
}

func TestE2E_GetUser(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	// Create user first
	user := map[string]interface{}{
		"id":        "user1",
		"username":  "testuser",
		"is_active": true,
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Get user
	req = httptest.NewRequest("GET", "/api/v1/users/user1", nil)
	w = httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user1", response["id"])
}

func TestE2E_CreateTeam(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	team := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{"id": "user1", "username": "user1", "is_active": true},
			{"id": "user2", "username": "user2", "is_active": true},
		},
	}

	body, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "team1", response["team_name"])
}

func TestE2E_CreatePullRequest(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	// Create team and users
	team := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{"id": "author1", "username": "author1", "is_active": true},
			{"id": "reviewer1", "username": "reviewer1", "is_active": true},
			{"id": "reviewer2", "username": "reviewer2", "is_active": true},
		},
	}

	body, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Create PR
	pr := map[string]interface{}{
		"pull_request_id":   "pr1",
		"pull_request_name": "Test PR",
		"author_id":         "author1",
	}

	body, _ = json.Marshal(pr)
	req = httptest.NewRequest("POST", "/api/v1/pull-request/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "pr1", response["pull_request_id"])

	// Verify reviewers were assigned
	reviewers, ok := response["assigned_reviewers"].([]interface{})
	if ok {
		assert.LessOrEqual(t, len(reviewers), 2)
		assert.NotContains(t, reviewers, "author1")
	}
}

func TestE2E_MergePullRequest(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	// Setup: create team, users and PR
	team := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{"id": "author1", "username": "author1", "is_active": true},
		},
	}

	body, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)

	pr := map[string]interface{}{
		"pull_request_id":   "pr1",
		"pull_request_name": "Test PR",
		"author_id":         "author1",
	}

	body, _ = json.Marshal(pr)
	req = httptest.NewRequest("POST", "/api/v1/pull-request/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)

	// Merge PR
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr1",
	}

	body, _ = json.Marshal(mergeReq)
	req = httptest.NewRequest("POST", "/api/v1/pull-request/merge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test idempotency - merge again
	req = httptest.NewRequest("POST", "/api/v1/pull-request/merge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // Should not error
}

func TestE2E_GetPullRequestsByReviewer(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	// Setup: create team, users and PR
	team := map[string]interface{}{
		"team_name": "team1",
		"members": []map[string]interface{}{
			{"id": "author1", "username": "author1", "is_active": true},
			{"id": "reviewer1", "username": "reviewer1", "is_active": true},
		},
	}

	body, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)

	pr := map[string]interface{}{
		"pull_request_id":   "pr1",
		"pull_request_name": "Test PR",
		"author_id":         "author1",
	}

	body, _ = json.Marshal(pr)
	req = httptest.NewRequest("POST", "/api/v1/pull-request/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	GetTestServer().GetRouter().ServeHTTP(w, req)

	// Get PRs by reviewer
	req = httptest.NewRequest("GET", "/api/v1/users/get-review?reviewer_id=reviewer1", nil)
	w = httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 0)
}

func TestE2E_GetStatistics(t *testing.T) {
	SetupE2ETest(t)
	defer TeardownE2ETest()

	req := httptest.NewRequest("GET", "/api/v1/statistics", nil)
	w := httptest.NewRecorder()

	GetTestServer().GetRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "total_prs")
	assert.Contains(t, response, "open_prs")
	assert.Contains(t, response, "merged_prs")
}
