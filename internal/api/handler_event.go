package api

import (
	"context"
	"errors"
	"fmt"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	"github.com/atcdot/SharedPrep/api/event/v1/eventv1connect"
	"github.com/atcdot/SharedPrep/internal/service"

	"connectrpc.com/connect"
)

type EventHandler struct {
	events       *service.EventService
	participants *service.ParticipantService
}

func NewEventHandler(events *service.EventService, participants *service.ParticipantService) *EventHandler {
	return &EventHandler{events: events, participants: participants}
}

func (h *EventHandler) CreateEvent(ctx context.Context, req *connect.Request[eventv1.CreateEventRequest]) (*connect.Response[eventv1.CreateEventResponse], error) {
	userID := UserIDFromContext(ctx)
	resp, err := h.events.CreateEvent(ctx, req.Msg, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (h *EventHandler) GetEvent(ctx context.Context, req *connect.Request[eventv1.GetEventRequest]) (*connect.Response[eventv1.GetEventResponse], error) {
	resp, err := h.events.GetEvent(ctx, req.Msg)
	if err != nil {
		if isNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	userID := UserIDFromContext(ctx)
	if userID != "" {
		pid, _ := h.participants.FindParticipantByUserAndEvent(ctx, userID, resp.Event.Id)
		resp.MyParticipantId = pid
	}

	return connect.NewResponse(resp), nil
}

func (h *EventHandler) UpdateEvent(ctx context.Context, req *connect.Request[eventv1.UpdateEventRequest]) (*connect.Response[eventv1.UpdateEventResponse], error) {
	resp, err := h.events.UpdateEvent(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *EventHandler) DeleteEvent(ctx context.Context, req *connect.Request[eventv1.DeleteEventRequest]) (*connect.Response[eventv1.DeleteEventResponse], error) {
	resp, err := h.events.DeleteEvent(ctx, req.Msg)
	if err != nil {
		if isNotFound(err) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}
func (h *EventHandler) ListMyEvents(ctx context.Context, req *connect.Request[eventv1.ListMyEventsRequest]) (*connect.Response[eventv1.ListMyEventsResponse], error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("login required"))
	}

	events, err := h.events.ListMyEvents(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&eventv1.ListMyEventsResponse{Events: events}), nil
}

func isNotFound(err error) bool {
	return errors.Is(err, service.ErrNotFound)
}

var _ eventv1connect.EventServiceHandler = (*EventHandler)(nil)
