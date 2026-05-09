package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	"github.com/atcdot/SharedPrep/internal/api"
	"github.com/atcdot/SharedPrep/internal/auth"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/testutil"

	"connectrpc.com/connect"
)

func setupEventTest(t *testing.T) (*api.EventHandler, *service.EventService, *service.ParticipantService) {
	t.Helper()
	db := testutil.NewTestDB(t)
	eventSvc := service.NewEventService(db)
	partSvc := service.NewParticipantService(db)
	handler := api.NewEventHandler(eventSvc, partSvc)
	return handler, eventSvc, partSvc
}

func TestEventHandler_CreateEvent(t *testing.T) {
	handler, _, _ := setupEventTest(t)

	resp, err := handler.CreateEvent(context.Background(), connect.NewRequest(&eventv1.CreateEventRequest{
		Title:       "Шашлыки",
		Description: ptrStr("На даче"),
	}))
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	if resp.Msg.Event.Title != "Шашлыки" {
		t.Errorf("title = %q, want %q", resp.Msg.Event.Title, "Шашлыки")
	}
	if resp.Msg.Event.ShareCode == "" {
		t.Error("expected non-empty share_code")
	}
}

func TestEventHandler_GetEvent(t *testing.T) {
	handler, eventSvc, _ := setupEventTest(t)
	ctx := context.Background()

	created, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	resp, err := handler.GetEvent(ctx, connect.NewRequest(&eventv1.GetEventRequest{
		ShareCode: created.Event.ShareCode,
	}))
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if resp.Msg.Event.Title != "Тест" {
		t.Errorf("title = %q, want %q", resp.Msg.Event.Title, "Тест")
	}
}

func TestEventHandler_GetEvent_NotFound(t *testing.T) {
	handler, _, _ := setupEventTest(t)

	_, err := handler.GetEvent(context.Background(), connect.NewRequest(&eventv1.GetEventRequest{
		ShareCode: "NOPE",
	}))
	if err == nil {
		t.Fatal("expected error for non-existent share_code")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("code = %v, want NotFound", connect.CodeOf(err))
	}
}

func TestEventHandler_UpdateEvent(t *testing.T) {
	handler, eventSvc, _ := setupEventTest(t)
	ctx := context.Background()

	created, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	resp, err := handler.UpdateEvent(ctx, connect.NewRequest(&eventv1.UpdateEventRequest{
		EventId: created.Event.Id,
		Title:   "Новое",
	}))
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if resp.Msg.Event.Title != "Новое" {
		t.Errorf("title = %q, want %q", resp.Msg.Event.Title, "Новое")
	}
}

func TestEventHandler_DeleteEvent(t *testing.T) {
	handler, eventSvc, _ := setupEventTest(t)
	ctx := context.Background()

	created, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	_, err := handler.DeleteEvent(ctx, connect.NewRequest(&eventv1.DeleteEventRequest{
		EventId: created.Event.Id,
	}))
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}

	_, err = handler.GetEvent(ctx, connect.NewRequest(&eventv1.GetEventRequest{
		ShareCode: created.Event.ShareCode,
	}))
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestEventHandler_DeleteEvent_NotFound(t *testing.T) {
	handler, _, _ := setupEventTest(t)

	_, err := handler.DeleteEvent(context.Background(), connect.NewRequest(&eventv1.DeleteEventRequest{
		EventId: "00000000-0000-0000-0000-000000000000",
	}))
	if err == nil {
		t.Fatal("expected error for non-existent event")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("code = %v, want NotFound", connect.CodeOf(err))
	}
}

func TestAuthMiddleware_JWT(t *testing.T) {
	secret := []byte("test-secret-key-32-chars-long-xxxxx")

	token, err := auth.GenerateToken("user-123", secret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	db := testutil.NewTestDB(t)
	partSvc := service.NewParticipantService(db)
	mw := api.NewAuthMiddleware(partSvc, secret, nil)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		uid := api.UserIDFromRequest(r)
		if uid != "user-123" {
			t.Errorf("user_id = %q, want %q", uid, "user-123")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rec := httptest.NewRecorder()

	mw.Wrap(next).ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func ptrStr(s string) *string { return &s }
