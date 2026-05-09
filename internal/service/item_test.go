package service_test

import (
	"context"
	"testing"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	participantv1 "github.com/atcdot/SharedPrep/api/participant/v1"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/testutil"
)

func setupItemTest(t *testing.T) (*service.EventService, *service.ParticipantService, *service.ItemService, string, string) {
	t.Helper()
	db := testutil.NewTestDB(t)
	eventSvc := service.NewEventService(db)
	partSvc := service.NewParticipantService(db)
	itemSvc := service.NewItemService(db)
	ctx := context.Background()

	ev, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")
	joined, _ := partSvc.JoinEvent(ctx, &participantv1.JoinEventRequest{
	}, "")

	return eventSvc, partSvc, itemSvc, ev.Event.Id, joined.Participant.Id
}

func TestCreateItem(t *testing.T) {
	_, _, itemSvc, eventID, participantID := setupItemTest(t)
	ctx := context.Background()

	item, err := itemSvc.CreateItem(ctx, participantID, eventID, "Мясо", 3)
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if item.Title != "Мясо" {
		t.Errorf("title = %q, want %q", item.Title, "Мясо")
	}
	if item.Quantity != 3 {
		t.Errorf("quantity = %d, want 3", item.Quantity)
	}
	if item.CreatorName != "Алексей" {
		t.Errorf("creator_name = %q, want %q", item.CreatorName, "Алексей")
	}
}

func TestUpdateItem(t *testing.T) {
	_, _, itemSvc, eventID, participantID := setupItemTest(t)
	ctx := context.Background()

	item, _ := itemSvc.CreateItem(ctx, participantID, eventID, "Мясо", 3)

	updated, err := itemSvc.UpdateItem(ctx, item.Id, "Курица", 5)
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if updated.Title != "Курица" {
		t.Errorf("title = %q, want %q", updated.Title, "Курица")
	}
	if updated.Quantity != 5 {
		t.Errorf("quantity = %d, want 5", updated.Quantity)
	}
}

func TestDeleteItem(t *testing.T) {
	_, _, itemSvc, eventID, participantID := setupItemTest(t)
	ctx := context.Background()

	item, _ := itemSvc.CreateItem(ctx, participantID, eventID, "Мясо", 3)

	err := itemSvc.DeleteItem(ctx, item.Id)
	if err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}

	items, _ := itemSvc.ListItems(ctx, eventID)
	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}

func TestClaimUnclaimItem(t *testing.T) {
	_, _, itemSvc, eventID, participantID := setupItemTest(t)
	ctx := context.Background()

	item, _ := itemSvc.CreateItem(ctx, participantID, eventID, "Мясо", 3)

	claimed, err := itemSvc.ClaimItem(ctx, item.Id, participantID)
	if err != nil {
		t.Fatalf("ClaimItem: %v", err)
	}
	if claimed.AssignedParticipantId == nil || *claimed.AssignedParticipantId != participantID {
		t.Errorf("assigned_participant_id = %v, want %q", claimed.AssignedParticipantId, participantID)
	}
	if claimed.AssignedParticipantName == nil || *claimed.AssignedParticipantName != "Алексей" {
		t.Errorf("assigned_participant_name = %v, want %q", claimed.AssignedParticipantName, "Алексей")
	}

	unclaimed, err := itemSvc.UnclaimItem(ctx, item.Id)
	if err != nil {
		t.Fatalf("UnclaimItem: %v", err)
	}
	if unclaimed.AssignedParticipantId != nil {
		t.Errorf("expected nil assigned_participant_id after unclaim, got %v", unclaimed.AssignedParticipantId)
	}
}

func TestListItems(t *testing.T) {
	_, _, itemSvc, eventID, participantID := setupItemTest(t)
	ctx := context.Background()

	itemSvc.CreateItem(ctx, participantID, eventID, "Мясо", 3)
	itemSvc.CreateItem(ctx, participantID, eventID, "Уголь", 2)

	items, err := itemSvc.ListItems(ctx, eventID)
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
}

func TestDeleteItem_NotFound(t *testing.T) {
	_, _, itemSvc, _, _ := setupItemTest(t)
	ctx := context.Background()

	err := itemSvc.DeleteItem(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("expected error for non-existent item")
	}
}

func TestClaimItem_NotFound(t *testing.T) {
	_, _, itemSvc, _, participantID := setupItemTest(t)
	ctx := context.Background()

	_, err := itemSvc.ClaimItem(ctx, "00000000-0000-0000-0000-000000000000", participantID)
	if err == nil {
		t.Error("expected error for non-existent item")
	}
}
