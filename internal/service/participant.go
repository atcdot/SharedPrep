package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	participantv1 "github.com/atcdot/SharedPrep/api/participant/v1"
	"github.com/atcdot/SharedPrep/internal/storage"

	"github.com/jackc/pgx/v5"
)

type ParticipantService struct {
	db *storage.Postgres
}

func NewParticipantService(db *storage.Postgres) *ParticipantService {
	return &ParticipantService{db: db}
}

func (s *ParticipantService) JoinEvent(ctx context.Context, req *participantv1.JoinEventRequest, userID string) (*participantv1.JoinEventResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	var eventID string
	err := s.db.Pool.QueryRow(ctx, `SELECT id FROM events WHERE share_code = $1`, req.ShareCode).Scan(&eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("find event: %w", err)
	}

	// Idempotency: if user already joined, return existing participant
	var existingID string
	err = s.db.Pool.QueryRow(ctx, `
		SELECT id FROM participants WHERE user_id = $1 AND event_id = $2
	`, userID, eventID).Scan(&existingID)
	if err == nil {
		return s.getParticipantByID(ctx, existingID)
	}

	var id, evID string
	var isAuthor bool
	var createdAt time.Time
	var displayName *string

	err = s.db.Pool.QueryRow(ctx, `
		INSERT INTO participants (event_id, is_author, user_id)
		VALUES ($1, FALSE, $2)
		RETURNING p.id, p.event_id, p.is_author, p.created_at,
		          (SELECT COALESCE(u.display_name, u.first_name) FROM users u WHERE u.id = p.user_id)
	`, eventID, userID).Scan(&id, &evID, &isAuthor, &createdAt, &displayName)
	if err != nil {
		return nil, fmt.Errorf("insert participant: %w", err)
	}

	return &participantv1.JoinEventResponse{
		Participant: &participantv1.Participant{
			Id: id, EventId: evID,
			IsAuthor: isAuthor, CreatedAt: createdAt.Format(time.RFC3339),
			DisplayName: displayName,
		},
	}, nil
}

func (s *ParticipantService) getParticipantByID(ctx context.Context, id string) (*participantv1.JoinEventResponse, error) {
	var evID string
	var isAuthor bool
	var createdAt time.Time
	var displayName *string
	var tgUsername *string
	err := s.db.Pool.QueryRow(ctx, `
		SELECT p.event_id, p.is_author, p.created_at,
		       u.display_name,
		       CASE WHEN u.show_telegram THEN u.username ELSE NULL END
		FROM participants p JOIN users u ON u.id = p.user_id
		WHERE p.id = $1
	`, id).Scan(&evID, &isAuthor, &createdAt, &displayName, &tgUsername)
	if err != nil {
		return nil, err
	}
	return &participantv1.JoinEventResponse{
		Participant: &participantv1.Participant{
			Id: id, EventId: evID,
			IsAuthor: isAuthor, CreatedAt: createdAt.Format(time.RFC3339),
			DisplayName: displayName, TelegramUsername: tgUsername,
		},
	}, nil
}
func (s *ParticipantService) FindParticipantByUserAndEvent(ctx context.Context, userID, eventID string) (string, error) {
	var id string
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id FROM participants WHERE user_id = $1 AND event_id = $2
	`, userID, eventID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *ParticipantService) ListParticipants(ctx context.Context, req *participantv1.ListParticipantsRequest) (*participantv1.ListParticipantsResponse, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT p.id, p.event_id, p.is_author, p.created_at,
		       u.display_name,
		       CASE WHEN u.show_telegram THEN u.username ELSE NULL END
		FROM participants p JOIN users u ON u.id = p.user_id
		WHERE p.event_id = $1 ORDER BY p.created_at`, req.EventId)
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	defer rows.Close()

	var participants []*participantv1.Participant
	for rows.Next() {
		var id, evID string
		var isAuthor bool
		var createdAt time.Time
		var displayName *string
		var tgUsername *string

		if err := rows.Scan(&id, &evID, &isAuthor, &createdAt, &displayName, &tgUsername); err != nil {
			return nil, fmt.Errorf("scan participant: %w", err)
		}
		participants = append(participants, &participantv1.Participant{
			Id: id, EventId: evID,
			IsAuthor: isAuthor, CreatedAt: createdAt.Format(time.RFC3339),
			DisplayName: displayName, TelegramUsername: tgUsername,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return &participantv1.ListParticipantsResponse{Participants: participants}, nil
}

func (s *ParticipantService) IsAuthor(ctx context.Context, participantID string) (bool, error) {
	var isAuthor bool
	err := s.db.Pool.QueryRow(ctx, `
		SELECT is_author FROM participants WHERE id = $1
	`, participantID).Scan(&isAuthor)
	if err != nil {
		return false, err
	}
	return isAuthor, nil
}

