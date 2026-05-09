package api

import (
	"context"
	"fmt"

	itemv1 "github.com/atcdot/SharedPrep/api/item/v1"
	"github.com/atcdot/SharedPrep/api/item/v1/itemv1connect"
	"github.com/atcdot/SharedPrep/internal/service"

	"connectrpc.com/connect"
)

type ItemHandler struct {
	items        *service.ItemService
	participants *service.ParticipantService
}

func NewItemHandler(items *service.ItemService, participants *service.ParticipantService) *ItemHandler {
	return &ItemHandler{items: items, participants: participants}
}
func (h *ItemHandler) resolveParticipantID(ctx context.Context, eventID string) (string, error) {
	uid := UserIDFromContext(ctx)
	if uid == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authenticated"))
	}
	pid, err := h.participants.FindParticipantByUserAndEvent(ctx, uid, eventID)
	if err != nil {
		return "", connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not a participant"))
	}
	return pid, nil
}


func (h *ItemHandler) CreateItem(ctx context.Context, req *connect.Request[itemv1.CreateItemRequest]) (*connect.Response[itemv1.CreateItemResponse], error) {
	participantID, err := h.resolveParticipantID(ctx, req.Msg.EventId)
	if err != nil {
		return nil, err
	}

	item, err := h.items.CreateItem(ctx, participantID, req.Msg.EventId, req.Msg.Title, req.Msg.Quantity)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.CreateItemResponse{Item: item}), nil
}

func (h *ItemHandler) UpdateItem(ctx context.Context, req *connect.Request[itemv1.UpdateItemRequest]) (*connect.Response[itemv1.UpdateItemResponse], error) {
	item, err := h.items.UpdateItem(ctx, req.Msg.ItemId, req.Msg.Title, req.Msg.Quantity)
	if err != nil {
		if err.Error() == "item not found" {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.UpdateItemResponse{Item: item}), nil
}

func (h *ItemHandler) DeleteItem(ctx context.Context, req *connect.Request[itemv1.DeleteItemRequest]) (*connect.Response[itemv1.DeleteItemResponse], error) {
	if err := h.items.DeleteItem(ctx, req.Msg.ItemId); err != nil {
		if err.Error() == "item not found" {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.DeleteItemResponse{}), nil
}

func (h *ItemHandler) ClaimItem(ctx context.Context, req *connect.Request[itemv1.ClaimItemRequest]) (*connect.Response[itemv1.ClaimItemResponse], error) {
	uid := UserIDFromContext(ctx)
	if uid == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authenticated"))
	}

	eventID, err := h.items.GetEventIDByItem(ctx, req.Msg.ItemId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("item not found"))
	}

	participantID, err := h.participants.FindParticipantByUserAndEvent(ctx, uid, eventID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not a participant"))
	}

	item, err := h.items.ClaimItem(ctx, req.Msg.ItemId, participantID)
	if err != nil {
		if err.Error() == "item not found" {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.ClaimItemResponse{Item: item}), nil
}

func (h *ItemHandler) UnclaimItem(ctx context.Context, req *connect.Request[itemv1.UnclaimItemRequest]) (*connect.Response[itemv1.UnclaimItemResponse], error) {
	item, err := h.items.UnclaimItem(ctx, req.Msg.ItemId)
	if err != nil {
		if err.Error() == "item not found" {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.UnclaimItemResponse{Item: item}), nil
}
func (h *ItemHandler) AssignItem(ctx context.Context, req *connect.Request[itemv1.AssignItemRequest]) (*connect.Response[itemv1.AssignItemResponse], error) {
	// Resolve caller's participant ID
	eventID, err := h.items.GetEventIDByItem(ctx, req.Msg.ItemId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("item not found"))
	}

	callerPID, err := h.resolveParticipantID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	isAuthor, err := h.participants.IsAuthor(ctx, callerPID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !isAuthor {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("only the event author can assign items"))
	}

	item, err := h.items.ClaimItem(ctx, req.Msg.ItemId, req.Msg.ParticipantId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.AssignItemResponse{Item: item}), nil
}


func (h *ItemHandler) ListItems(ctx context.Context, req *connect.Request[itemv1.ListItemsRequest]) (*connect.Response[itemv1.ListItemsResponse], error) {
	items, err := h.items.ListItems(ctx, req.Msg.EventId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&itemv1.ListItemsResponse{Items: items}), nil
}

var _ itemv1connect.ItemServiceHandler = (*ItemHandler)(nil)
