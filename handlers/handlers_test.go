package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EgorSheff/go-project-278/db/generated"
	"github.com/EgorSheff/go-project-278/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
)

func setupRouter(t *testing.T) (*gin.Engine, pgxmock.PgxPoolIface) {
	t.Helper()
	t.Setenv("BASE_URL", "http://localhost")

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	dao := generated.New(mock)
	handlers.RegisterHandlers(r, dao)

	return r, mock
}

func TestRedirect_Success(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, original_url, short_name from links WHERE short_name").
		WithArgs("abc").
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}).
			AddRow(int32(1), "https://example.com", "abc"))

	mock.ExpectQuery("INSERT INTO visits").
		WithArgs(int32(1), pgxmock.AnyArg(), pgxmock.AnyArg(), int16(http.StatusTemporaryRedirect)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "link_id", "created_at", "ip", "user_agent", "status"}).
			AddRow(int32(1), int32(1), pgtype.Timestamp{Time: time.Now(), Valid: true}, "127.0.0.1", "", int16(307)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/r/abc", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("expected Location https://example.com, got %s", loc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, original_url, short_name from links WHERE short_name").
		WithArgs("unknown").
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/r/unknown", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetLinks(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, original_url, short_name FROM links ORDER BY id").
		WithArgs(int32(100), int32(0)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}).
			AddRow(int32(1), "https://example.com", "abc").
			AddRow(int32(2), "https://google.com", "def"))

	mock.ExpectQuery("SELECT count").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/links", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var links []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &links); err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCreateLink(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("INSERT INTO links").
		WithArgs("https://example.com", "mycode").
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}).
			AddRow(int32(1), "https://example.com", "mycode"))

	body, _ := json.Marshal(map[string]string{
		"original_url": "https://example.com",
		"short_name":   "mycode",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/links", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["short_name"] != "mycode" {
		t.Errorf("expected short_name mycode, got %v", resp["short_name"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCreateLink_ValidationError(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	body, _ := json.Marshal(map[string]string{
		"original_url": "not-a-url",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/links", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetLink(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, original_url, short_name FROM links WHERE id").
		WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}).
			AddRow(int32(1), "https://example.com", "abc"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/links/1", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetLink_NotFound(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, original_url, short_name FROM links WHERE id").
		WithArgs(int32(999)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/links/999", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDeleteLink(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectExec("DELETE FROM links WHERE id").
		WithArgs(int32(1)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "DELETE", "/api/links/1", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUpdateLink(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	mock.ExpectQuery("UPDATE links SET").
		WithArgs("https://updated.com", "newcode", int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "original_url", "short_name"}).
			AddRow(int32(1), "https://updated.com", "newcode"))

	body, _ := json.Marshal(map[string]string{
		"original_url": "https://updated.com",
		"short_name":   "newcode",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "PUT", "/api/links/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetLinkVisits(t *testing.T) {
	r, mock := setupRouter(t)
	defer mock.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT id, link_id, created_at, ip, user_agent, status FROM visits ORDER BY id").
		WithArgs(int32(100), int32(0)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "link_id", "created_at", "ip", "user_agent", "status"}).
			AddRow(int32(1), int32(1), pgtype.Timestamp{Time: now, Valid: true}, "192.168.1.1", "Mozilla/5.0", int16(307)))

	mock.ExpectQuery("SELECT count").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/link_visits", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var visits []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &visits); err != nil {
		t.Fatal(err)
	}
	if len(visits) != 1 {
		t.Errorf("expected 1 visit, got %d", len(visits))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
