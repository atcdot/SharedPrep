package api_test

import (
	"context"
	"testing"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	itemv1 "github.com/atcdot/SharedPrep/api/item/v1"
	participantv1 "github.com/atcdot/SharedPrep/api/participant/v1"
	"github.com/atcdot/SharedPrep/internal/api"
	"github.com/atcdot/SharedPrep/internal/service"
	"github.com/atcdot/SharedPrep/internal/testutil"

	"connectrpc.com/connect"
)

func setupItemHandlerTest(t *testing.T) (*api.ItemHandler, string, string) {
	t.Helper()
	db := testutil.NewTestDB(t)
	eventSvc := service.NewEventService(db)
	partSvc := service.NewParticipantService(db)
	itemSvc := service.NewItemService(db)
	handler := api.NewItemHandler(itemSvc, partSvc)
	ctx := context.Background()

	ev, _ := eventSvc.CreateEvent(ctx, &eventv1.CreateEventRequest{
	}, "")
	joined, _ := partSvc.JoinEvent(ctx, &participantv1.JoinEventRequest{
	}, "")

	return handler, ev.Event.Id, joined.Participant.Id
}

func ctxWithParticipant(participantID string) context.Context {
	return api.WithParticipantID(context.Background(), participantID)
}

func TestItemHandler_CreateItem(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)

	resp, err := handler.CreateItem(ctxWithParticipant(participantID), connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID,
		Title:   "Мясо",
		Quantity: 3,
	}))
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if resp.Msg.Item.Title != "Мясо" {
		t.Errorf("title = %q, want %q", resp.Msg.Item.Title, "Мясо")
	}
	if resp.Msg.Item.Quantity != 3 {
		t.Errorf("quantity = %d, want 3", resp.Msg.Item.Quantity)
	}
}

func TestItemHandler_CreateItem_Unauthenticated(t *testing.T) {
	handler, eventID, _ := setupItemHandlerTest(t)

	_, err := handler.CreateItem(context.Background(), connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID,
		Title:   "Мясо",
		Quantity: 1,
	}))
	if err == nil {
		t.Fatal("expected error when unauthenticated")
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want Unauthenticated", connect.CodeOf(err))
	}
}

func TestItemHandler_UpdateItem(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)
	ctx := ctxWithParticipant(participantID)

	created, _ := handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Мясо", Quantity: 3,
	}))

	resp, err := handler.UpdateItem(ctx, connect.NewRequest(&itemv1.UpdateItemRequest{
		ItemId:  created.Msg.Item.Id,
		Title:   "Курица",
		Quantity: 5,
	}))
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if resp.Msg.Item.Title != "Курица" {
		t.Errorf("title = %q, want %q", resp.Msg.Item.Title, "Курица")
	}
}

func TestItemHandler_DeleteItem(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)
	ctx := ctxWithParticipant(participantID)

	created, _ := handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Мясо", Quantity: 3,
	}))

	_, err := handler.DeleteItem(ctx, connect.NewRequest(&itemv1.DeleteItemRequest{
		ItemId: created.Msg.Item.Id,
	}))
	if err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
}

func TestItemHandler_ClaimItem(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)
	ctx := ctxWithParticipant(participantID)

	created, _ := handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Мясо", Quantity: 3,
	}))

	resp, err := handler.ClaimItem(ctxWithParticipant(participantID), connect.NewRequest(&itemv1.ClaimItemRequest{
		ItemId: created.Msg.Item.Id,
	}))
	if err != nil {
		t.Fatalf("ClaimItem: %v", err)
	}
	if resp.Msg.Item.AssignedParticipantId == nil {
		t.Error("expected assigned_participant_id after claim")
	}
}

func TestItemHandler_ClaimItem_Unauthenticated(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)
	ctx := ctxWithParticipant(participantID)

	created, _ := handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Мясо", Quantity: 3,
	}))

	_, err := handler.ClaimItem(context.Background(), connect.NewRequest(&itemv1.ClaimItemRequest{
		ItemId: created.Msg.Item.Id,
	}))
	if err == nil {
		t.Fatal("expected error when unauthenticated")
	}
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want Unauthenticated", connect.CodeOf(err))
	}
}

func TestItemHandler_UnclaimItem(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)
	ctx := ctxWithParticipant(participantID)

	created, _ := handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Мясо", Quantity: 3,
	}))
	handler.ClaimItem(ctxWithParticipant(participantID), connect.NewRequest(&itemv1.ClaimItemRequest{
		ItemId: created.Msg.Item.Id,
	}))

	resp, err := handler.UnclaimItem(ctx, connect.NewRequest(&itemv1.UnclaimItemRequest{
		ItemId: created.Msg.Item.Id,
	}))
	if err != nil {
		t.Fatalf("UnclaimItem: %v", err)
	}
	if resp.Msg.Item.AssignedParticipantId != nil {
		t.Error("expected nil assigned_participant_id after unclaim")
	}
}

func TestItemHandler_ListItems(t *testing.T) {
	handler, eventID, participantID := setupItemHandlerTest(t)
	ctx := ctxWithParticipant(participantID)

	handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Мясо", Quantity: 3,
	}))
	handler.CreateItem(ctx, connect.NewRequest(&itemv1.CreateItemRequest{
		EventId: eventID, Title: "Уголь", Quantity: 2,
	}))

	resp, err := handler.ListItems(ctx, connect.NewRequest(&itemv1.ListItemsRequest{
		EventId: eventID,
	}))
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if len(resp.Msg.Items) != 2 {
		t.Errorf("got %d items, want 2", len(resp.Msg.Items))
	}
}
