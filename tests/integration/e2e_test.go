//go:build integration

package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/university/sports-event-planner-platform/pkg/grpcx"
	authv1 "github.com/university/sports-event-planner-platform/services/auth-service/proto/auth/v1"
)

const (
	apiBase   = "http://127.0.0.1:18080"
	jwtSecret = "integration-test-secret"
)

func TestBackendE2E(t *testing.T) {
	root := repoRoot(t)
	project := "sports-e2e-" + randomHex(t, 4)
	compose := composeRunner{t: t, root: root, project: project}

	t.Cleanup(func() {
		compose.run("down", "-v", "--remove-orphans")
	})
	compose.run("up", "--build", "-d")

	waitReady(t)

	owner := registerHTTP(t, "owner-"+randomHex(t, 4)+"@example.com", "password123")
	login := loginHTTP(t, owner.Email, "password123")
	if login.UserId != owner.UserId || login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatalf("unexpected login response: %#v", login)
	}
	refreshed := refreshHTTP(t, owner.RefreshToken)
	if refreshed.UserId != owner.UserId || refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatalf("unexpected refresh response: %#v", refreshed)
	}

	expectAPIError(t, http.MethodPost, "/v1/auth/login", "", map[string]any{
		"email":    owner.Email,
		"password": "wrong-password",
	}, http.StatusUnauthorized, "INVALID_CREDENTIALS")
	expectAPIError(t, http.MethodGet, "/v1/auth/me", "bad-token", nil, http.StatusUnauthorized, "INVALID_TOKEN")
	expectAPIError(t, http.MethodGet, "/v1/auth/me", "", nil, http.StatusUnauthorized, "UNAUTHORIZED")
	expectAPIErrorWithToken(t, http.MethodGet, "/v1/auth/me", expiredToken(t, owner.UserId, owner.Email), nil, http.StatusUnauthorized, "INVALID_TOKEN")

	testAuthGRPC(t)

	expectAPIError(t, http.MethodPost, "/v1/events", owner.AccessToken, map[string]any{
		"sport":      "football",
		"start_time": "2026-06-01T10:00:00Z",
	}, http.StatusBadRequest, "INVALID_REQUEST")

	event := createEventHTTP(t, owner.AccessToken, map[string]any{
		"sport":            "football",
		"category":         "intramural",
		"competition":      "campus cup",
		"title":            "Integration Match",
		"description":      "integration flow",
		"location":         "Main Field",
		"start_time":       "2026-06-01T10:00:00Z",
		"end_time":         "2026-06-01T12:00:00Z",
		"country":          "kz",
		"city":             "almaty",
		"max_participants": float64(1),
	})
	if event.CreatorId != owner.UserId || event.Id == "" {
		t.Fatalf("unexpected created event: %#v", event)
	}

	created := getEventHTTP(t, owner.AccessToken, event.Id)
	if created.Title != "Integration Match" {
		t.Fatalf("unexpected get event title: %q", created.Title)
	}
	if exists := redisExists(t, compose, "event:"+event.Id); exists != 1 {
		t.Fatalf("expected event cache to be populated, got EXISTS=%d", exists)
	}

	events := listEventsHTTP(t, owner.AccessToken, "sport=football&country=kz&page=1&page_size=5")
	if len(events) == 0 {
		t.Fatal("expected list events to return at least one event")
	}

	other := registerHTTP(t, "other-"+randomHex(t, 4)+"@example.com", "password123")
	expectAPIError(t, http.MethodPut, "/v1/events/"+event.Id, other.AccessToken, updatePayload("Other Title", 1), http.StatusForbidden, "FORBIDDEN")

	joined := joinEventHTTP(t, owner.AccessToken, event.Id)
	if joined.ParticipantsCount != 1 {
		t.Fatalf("expected participants_count=1 after join, got %d", joined.ParticipantsCount)
	}
	expectAPIError(t, http.MethodPost, "/v1/events/"+event.Id+"/join", other.AccessToken, map[string]any{}, http.StatusConflict, "EVENT_FULL")

	left := leaveEventHTTP(t, owner.AccessToken, event.Id)
	if left.ParticipantsCount != 0 {
		t.Fatalf("expected participants_count=0 after leave, got %d", left.ParticipantsCount)
	}

	updated := updateEventHTTP(t, owner.AccessToken, event.Id, updatePayload("Updated Match", 2))
	if updated.Title != "Updated Match" || updated.MaxParticipants != 2 {
		t.Fatalf("unexpected updated event: %#v", updated)
	}
	if exists := redisExists(t, compose, "event:"+event.Id); exists != 0 {
		t.Fatalf("expected event cache invalidated after update, got EXISTS=%d", exists)
	}
	updatedFromDB := getEventHTTP(t, owner.AccessToken, event.Id)
	if updatedFromDB.Title != "Updated Match" {
		t.Fatalf("expected updated title from DB, got %q", updatedFromDB.Title)
	}

	expectNotification(t, owner.AccessToken, owner.UserId, "New event: Integration Match")
	compose.run("restart", "nats")
	time.Sleep(5 * time.Second)
	reconnectedEvent := createEventHTTP(t, owner.AccessToken, map[string]any{
		"sport":            "football",
		"category":         "intramural",
		"competition":      "campus cup",
		"title":            "Reconnected Match",
		"description":      "nats reconnect flow",
		"location":         "Main Field",
		"start_time":       "2026-06-02T10:00:00Z",
		"end_time":         "2026-06-02T12:00:00Z",
		"country":          "kz",
		"city":             "almaty",
		"max_participants": float64(2),
	})
	expectNotification(t, owner.AccessToken, owner.UserId, "New event: Reconnected Match")
	deleteEventHTTP(t, owner.AccessToken, reconnectedEvent.Id)

	expectAPIError(t, http.MethodDelete, "/v1/events/"+event.Id, other.AccessToken, nil, http.StatusForbidden, "FORBIDDEN")
	deleteEventHTTP(t, owner.AccessToken, event.Id)
	expectAPIError(t, http.MethodGet, "/v1/events/"+event.Id, owner.AccessToken, nil, http.StatusNotFound, "EVENT_NOT_FOUND")
}

func testAuthGRPC(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		"127.0.0.1:15051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(grpcx.JSONCodecName)),
	)
	if err != nil {
		t.Fatalf("dial auth grpc: %v", err)
	}
	defer conn.Close()

	client := authv1.NewAuthServiceClient(conn)
	email := "grpc-" + randomHex(t, 4) + "@example.com"
	registered, err := client.Register(ctx, &authv1.RegisterRequest{Email: email, Password: "password123", Role: "student"})
	if err != nil {
		t.Fatalf("grpc register: %v", err)
	}
	loggedIn, err := client.Login(ctx, &authv1.LoginRequest{Email: email, Password: "password123"})
	if err != nil {
		t.Fatalf("grpc login: %v", err)
	}
	if loggedIn.UserId != registered.UserId {
		t.Fatalf("grpc user mismatch: register=%s login=%s", registered.UserId, loggedIn.UserId)
	}
	refreshed, err := client.RefreshToken(ctx, &authv1.RefreshTokenRequest{RefreshToken: loggedIn.RefreshToken})
	if err != nil {
		t.Fatalf("grpc refresh: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatalf("grpc refresh returned empty tokens: %#v", refreshed)
	}
	invalid, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: "not-a-token"})
	if err != nil {
		t.Fatalf("grpc invalid token should return response, got error: %v", err)
	}
	if invalid.Valid {
		t.Fatal("grpc invalid token unexpectedly valid")
	}
}

func registerHTTP(t *testing.T, email, password string) authSession {
	t.Helper()
	var out authv1.AuthResponse
	doJSON(t, http.MethodPost, "/v1/auth/register", "", map[string]any{
		"email":    email,
		"password": password,
		"role":     "student",
	}, http.StatusOK, &out)
	return authSession{AuthResponse: out, Email: email}
}

func loginHTTP(t *testing.T, email, password string) authv1.AuthResponse {
	t.Helper()
	var out authv1.AuthResponse
	doJSON(t, http.MethodPost, "/v1/auth/login", "", map[string]any{"email": email, "password": password}, http.StatusOK, &out)
	return out
}

func refreshHTTP(t *testing.T, refreshToken string) authv1.AuthResponse {
	t.Helper()
	var out authv1.AuthResponse
	doJSON(t, http.MethodPost, "/v1/auth/refresh", "", map[string]any{"refresh_token": refreshToken}, http.StatusOK, &out)
	return out
}

func createEventHTTP(t *testing.T, token string, payload map[string]any) eventDTO {
	t.Helper()
	var out eventResponse
	doJSON(t, http.MethodPost, "/v1/events", token, payload, http.StatusOK, &out)
	return out.Event
}

func getEventHTTP(t *testing.T, token string, id string) eventDTO {
	t.Helper()
	var out eventResponse
	doJSON(t, http.MethodGet, "/v1/events/"+id, token, nil, http.StatusOK, &out)
	return out.Event
}

func listEventsHTTP(t *testing.T, token string, query string) []eventDTO {
	t.Helper()
	var out listEventsResponse
	doJSON(t, http.MethodGet, "/v1/events?"+query, token, nil, http.StatusOK, &out)
	return out.Events
}

func updateEventHTTP(t *testing.T, token string, id string, payload map[string]any) eventDTO {
	t.Helper()
	var out eventResponse
	doJSON(t, http.MethodPut, "/v1/events/"+id, token, payload, http.StatusOK, &out)
	return out.Event
}

func joinEventHTTP(t *testing.T, token string, id string) eventDTO {
	t.Helper()
	var out eventResponse
	doJSON(t, http.MethodPost, "/v1/events/"+id+"/join", token, map[string]any{}, http.StatusOK, &out)
	return out.Event
}

func leaveEventHTTP(t *testing.T, token string, id string) eventDTO {
	t.Helper()
	var out eventResponse
	doJSON(t, http.MethodPost, "/v1/events/"+id+"/leave", token, map[string]any{}, http.StatusOK, &out)
	return out.Event
}

func deleteEventHTTP(t *testing.T, token string, id string) {
	t.Helper()
	var out map[string]any
	doJSON(t, http.MethodDelete, "/v1/events/"+id, token, nil, http.StatusOK, &out)
}

func expectNotification(t *testing.T, token string, userID string, subject string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var out notificationsResponse
		doJSON(t, http.MethodGet, "/v1/users/"+userID+"/notifications", token, nil, http.StatusOK, &out)
		for _, notification := range out.Notifications {
			if notification.Subject == subject {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("notification with subject %q was not delivered", subject)
}

func expectAPIError(t *testing.T, method, path, token string, payload any, statusCode int, code string) {
	t.Helper()
	if token == "" {
		var out apiError
		doJSON(t, method, path, "", payload, statusCode, &out)
		if out.Error.Code != code {
			t.Fatalf("expected error code %s, got %#v", code, out)
		}
		return
	}
	expectAPIErrorWithToken(t, method, path, token, payload, statusCode, code)
}

func expectAPIErrorWithToken(t *testing.T, method, path, token string, payload any, statusCode int, code string) {
	t.Helper()
	var out apiError
	doJSON(t, method, path, token, payload, statusCode, &out)
	if out.Error.Code != code {
		t.Fatalf("expected error code %s, got %#v", code, out)
	}
}

func doJSON(t *testing.T, method, path, token string, payload any, statusCode int, out any) {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request: %v", err)
		}
	}
	req, err := http.NewRequest(method, apiBase+path, &body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != statusCode {
		var raw bytes.Buffer
		_, _ = raw.ReadFrom(resp.Body)
		t.Fatalf("%s %s: expected status %d got %d body=%s", method, path, statusCode, resp.StatusCode, raw.String())
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
}

func waitReady(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(apiBase + "/readyz")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("api gateway did not become ready")
}

func redisExists(t *testing.T, compose composeRunner, key string) int {
	t.Helper()
	out := compose.output("exec", "-T", "redis", "redis-cli", "EXISTS", key)
	out = strings.TrimSpace(out)
	if out == "1" {
		return 1
	}
	if out == "0" {
		return 0
	}
	t.Fatalf("unexpected redis EXISTS output: %q", out)
	return 0
}

func updatePayload(title string, maxParticipants int) map[string]any {
	return map[string]any{
		"sport":            "football",
		"category":         "intramural",
		"competition":      "campus cup",
		"title":            title,
		"description":      "updated integration flow",
		"location":         "Main Field",
		"start_time":       "2026-06-01T10:00:00Z",
		"end_time":         "2026-06-01T12:00:00Z",
		"country":          "kz",
		"city":             "almaty",
		"max_participants": maxParticipants,
	}
}

func expiredToken(t *testing.T, userID string, email string) string {
	t.Helper()
	now := time.Now().Add(-2 * time.Hour)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    "student",
		"sub":     userID,
		"iat":     now.Unix(),
		"exp":     now.Add(time.Hour).Unix(),
	}).SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	return token
}

type composeRunner struct {
	t       *testing.T
	root    string
	project string
}

func (c composeRunner) run(args ...string) {
	c.t.Helper()
	_ = c.output(args...)
}

func (c composeRunner) output(args ...string) string {
	c.t.Helper()
	full := append([]string{"compose", "-f", "docker-compose.test.yml", "-p", c.project}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Dir = c.root
	raw, err := cmd.CombinedOutput()
	if err != nil {
		c.t.Fatalf("docker %s failed: %v\n%s", strings.Join(full, " "), err, string(raw))
	}
	return string(raw)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func randomHex(t *testing.T, n int) string {
	t.Helper()
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("random: %v", err)
	}
	return hex.EncodeToString(buf)
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type authSession struct {
	authv1.AuthResponse
	Email string
}

type eventResponse struct {
	Event eventDTO `json:"event"`
}

type listEventsResponse struct {
	Events []eventDTO `json:"events"`
}

type eventDTO struct {
	Id                string `json:"id"`
	CreatorId         string `json:"creator_id"`
	Sport             string `json:"sport"`
	Category          string `json:"category"`
	Competition       string `json:"competition"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Location          string `json:"location"`
	StartTime         string `json:"start_time"`
	EndTime           string `json:"end_time"`
	Status            string `json:"status"`
	Country           string `json:"country"`
	City              string `json:"city"`
	MaxParticipants   int32  `json:"max_participants"`
	ParticipantsCount int32  `json:"participants_count"`
}

type notificationsResponse struct {
	Notifications []struct {
		Id      string `json:"id"`
		UserId  string `json:"user_id"`
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Status  string `json:"status"`
	} `json:"notifications"`
}
