package service_test

import (
	"context"
	"testing"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/testutil"

	"github.com/google/uuid"
)

func TestCreateEvent(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewEventService(db)
	ctx := context.Background()

	resp, err := svc.CreateEvent(ctx, &eventv1.CreateEventRequest{
		Title:       "Шашлыки",
		Description: ptrStr("На даче"),
		Date:        ptrStr("2025-06-15"),
	}, "")
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}

	ev := resp.Event
	if ev.Id == "" {
		t.Error("expected non-empty id")
	}
	if ev.ShareCode == "" {
		t.Error("expected non-empty share_code")
	}
	if ev.Title != "Шашлыки" {
		t.Errorf("title = %q, want %q", ev.Title, "Шашлыки")
	}
	if ev.Description == nil || *ev.Description != "На даче" {
		t.Errorf("description = %v, want %q", ev.Description, "На даче")
	}
}

func TestGetEvent(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewEventService(db)
	ctx := context.Background()

	created, _ := svc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	resp, err := svc.GetEvent(ctx, &eventv1.GetEventRequest{
		ShareCode: created.Event.ShareCode,
	})
	if err != nil {
		t.Fatalf("GetEvent: %v", err)
	}
	if resp.Event.Title != "Тест" {
		t.Errorf("title = %q, want %q", resp.Event.Title, "Тест")
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewEventService(db)
	ctx := context.Background()

	_, err := svc.GetEvent(ctx, &eventv1.GetEventRequest{ShareCode: "NOPE"})
	if err == nil {
		t.Error("expected error for non-existent share_code")
	}
}

func TestUpdateEvent(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewEventService(db)
	ctx := context.Background()

	created, _ := svc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	resp, err := svc.UpdateEvent(ctx, &eventv1.UpdateEventRequest{
		EventId:     created.Event.Id,
		Title:       "Новое",
		Description: ptrStr("Обновлено"),
	})
	if err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	if resp.Event.Title != "Новое" {
		t.Errorf("title = %q, want %q", resp.Event.Title, "Новое")
	}
}

func TestDeleteEvent(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewEventService(db)
	ctx := context.Background()

	created, _ := svc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	_, err := svc.DeleteEvent(ctx, &eventv1.DeleteEventRequest{EventId: created.Event.Id})
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}

	_, err = svc.GetEvent(ctx, &eventv1.GetEventRequest{ShareCode: created.Event.ShareCode})
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewEventService(db)
	ctx := context.Background()

	_, err := svc.DeleteEvent(ctx, &eventv1.DeleteEventRequest{EventId: uuid.New().String()})
	if err == nil {
		t.Error("expected error for non-existent event")
	}
}

func ptrStr(s string) *string { return &s }
