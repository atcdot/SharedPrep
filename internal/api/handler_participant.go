package api

import (
	"context"

	participantv1 "github.com/atcdot/SharedPrep/api/participant/v1"
	"github.com/atcdot/SharedPrep/api/participant/v1/participantv1connect"
	"github.com/atcdot/SharedPrep/internal/service"

	"connectrpc.com/connect"
)

type ParticipantHandler struct {
	participants *service.ParticipantService
}

func NewParticipantHandler(participants *service.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{participants: participants}
}

func (h *ParticipantHandler) JoinEvent(ctx context.Context, req *connect.Request[participantv1.JoinEventRequest]) (*connect.Response[participantv1.JoinEventResponse], error) {
	userID := UserIDFromContext(ctx)
	resp, err := h.participants.JoinEvent(ctx, req.Msg, userID)
	if err != nil {
		if err.Error() == "event not found" {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (h *ParticipantHandler) ListParticipants(ctx context.Context, req *connect.Request[participantv1.ListParticipantsRequest]) (*connect.Response[participantv1.ListParticipantsResponse], error) {
	resp, err := h.participants.ListParticipants(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ participantv1connect.ParticipantServiceHandler = (*ParticipantHandler)(nil)
