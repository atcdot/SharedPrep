package service_test

import (
	"context"
	"testing"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	participantv1 "github.com/atcdot/SharedPrep/api/participant/v1"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/testutil"
)

func TestJoinEvent(t *testing.T) {
	db := testutil.NewTestDB(t)
	eventSvc := service.NewEventService(db)
	partSvc := service.NewParticipantService(db)
	ctx := context.Background()

	ev, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	resp, err := partSvc.JoinEvent(ctx, &participantv1.JoinEventRequest{
		ShareCode: ev.Event.ShareCode,
	}, "")
	if err != nil {
		t.Fatalf("JoinEvent: %v", err)
	}

	p := resp.Participant
	if p.Name != "Алексей" {
		t.Errorf("name = %q, want %q", p.Name, "Алексей")
	}
	if p.IsAuthor {
		t.Error("joined participant should not be author")
	}
}

func TestJoinEvent_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewParticipantService(db)
	ctx := context.Background()

	_, err := svc.JoinEvent(ctx, &participantv1.JoinEventRequest{
	}, "")
	if err == nil {
		t.Error("expected error for non-existent share_code")
	}
}

func TestListParticipants(t *testing.T) {
	db := testutil.NewTestDB(t)
	eventSvc := service.NewEventService(db)
	partSvc := service.NewParticipantService(db)
	ctx := context.Background()

	ev, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")
	partSvc.JoinEvent(ctx, &participantv1.JoinEventRequest{
	}, "")

	resp, err := partSvc.ListParticipants(ctx, &participantv1.ListParticipantsRequest{
		EventId: ev.Event.Id,
	})
	if err != nil {
		t.Fatalf("ListParticipants: %v", err)
	}
	if len(resp.Participants) != 2 {
		t.Errorf("got %d participants, want 2", len(resp.Participants))
	}
}

func TestResolveByToken_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := service.NewParticipantService(db)
	ctx := context.Background()

	_, err := svc.ResolveByToken(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}
