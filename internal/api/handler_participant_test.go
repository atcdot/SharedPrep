package api_test

import (
	"context"
	"testing"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	participantv1 "github.com/atcdot/SharedPrep/api/participant/v1"
	"github.com/atcdot/SharedPrep/internal/api"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/testutil"

	"connectrpc.com/connect"
)

func setupParticipantTest(t *testing.T) (*api.ParticipantHandler, *service.EventService, *service.ParticipantService) {
	t.Helper()
	db := testutil.NewTestDB(t)
	eventSvc := service.NewEventService(db)
	partSvc := service.NewParticipantService(db)
	handler := api.NewParticipantHandler(partSvc)
	return handler, eventSvc, partSvc
}

func TestParticipantHandler_JoinEvent(t *testing.T) {
	handler, eventSvc, _ := setupParticipantTest(t)
	ctx := context.Background()

	ev, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")

	resp, err := handler.JoinEvent(ctx, connect.NewRequest(&participantv1.JoinEventRequest{
		ShareCode: ev.Event.ShareCode,
	}))
	if err != nil {
		t.Fatalf("JoinEvent: %v", err)
	}

	if resp.Msg.Participant.Id == "" {
		t.Error("expected non-empty participant id")
	}
}

func TestParticipantHandler_JoinEvent_NotFound(t *testing.T) {
	handler, _, _ := setupParticipantTest(t)

	_, err := handler.JoinEvent(context.Background(), connect.NewRequest(&participantv1.JoinEventRequest{
	}))
	if err == nil {
		t.Fatal("expected error")
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("code = %v, want NotFound", connect.CodeOf(err))
	}
}

func TestParticipantHandler_ListParticipants(t *testing.T) {
	handler, eventSvc, partSvc := setupParticipantTest(t)
	ctx := context.Background()

	ev, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")
	partSvc.JoinEvent(ctx, &participantv1.JoinEventRequest{
	}, "")

	resp, err := handler.ListParticipants(ctx, connect.NewRequest(&participantv1.ListParticipantsRequest{
		EventId: ev.Event.Id,
	}))
	if err != nil {
		t.Fatalf("ListParticipants: %v", err)
	}
	if len(resp.Msg.Participants) != 2 {
		t.Errorf("got %d participants, want 2", len(resp.Msg.Participants))
	}
}
