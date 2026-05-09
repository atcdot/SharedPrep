package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	eventv1 "github.com/atcdot/SharedPrep/api/event/v1"
	"github.com/atcdot/SharedPrep/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EventService struct {
	db *storage.Postgres
}

func NewEventService(db *storage.Postgres) *EventService {
	return &EventService{db: db}
}

func (s *EventService) CreateEvent(ctx context.Context, req *eventv1.CreateEventRequest, userID string) (*eventv1.CreateEventResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("authentication required")
	}

	shareCode, err := generateShareCode()
	if err != nil {
		return nil, fmt.Errorf("generate share code: %w", err)
	}

	var id string
	var sc string
	var title string
	var desc *string
	var eventDate *time.Time
	var createdAt time.Time
	var link *string

	err = s.db.Pool.QueryRow(ctx, `
		INSERT INTO events (share_code, title, description, event_date, link)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, share_code, title, description, event_date, created_at, link`,
		shareCode, req.Title, req.Description, timePtr(req.Date), req.Link,
	).Scan(&id, &sc, &title, &desc, &eventDate, &createdAt, &link)
	if err != nil {
		return nil, fmt.Errorf("insert event: %w", err)
	}

	var dummy string
	err = s.db.Pool.QueryRow(ctx, `
		INSERT INTO participants (event_id, is_author, user_id)
		VALUES ($1, TRUE, $2)
		RETURNING id`,
		id, userID,
	).Scan(&dummy)
	if err != nil {
		return nil, fmt.Errorf("insert author: %w", err)
	}

	authorName, _ := s.userName(ctx, userID)

	return &eventv1.CreateEventResponse{Event: &eventv1.Event{
		Id: id, ShareCode: sc, Title: title, Description: desc,
		Date: formatTimePtr(eventDate), AuthorName: authorName,
		CreatedAt: createdAt.Format(time.RFC3339), Link: link,
	}}, nil
}

func (s *EventService) userName(ctx context.Context, userID string) (string, bool) {
	var name *string
	var fallback string
	err := s.db.Pool.QueryRow(ctx, `SELECT display_name, first_name FROM users WHERE id = $1`, userID).Scan(&name, &fallback)
	if err != nil {
		return "", false
	}
	if name != nil {
		return *name, true
	}
	return fallback, true
}
func (s *EventService) authorNameByEvent(ctx context.Context, eventID string) (string, bool) {
	var name string
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(u.display_name, u.first_name)
		FROM participants p JOIN users u ON u.id = p.user_id
		WHERE p.event_id = $1 AND p.is_author = TRUE`, eventID,
	).Scan(&name)
	if err != nil {
		return "", false
	}
	return name, true
}


func (s *EventService) GetEvent(ctx context.Context, req *eventv1.GetEventRequest) (*eventv1.GetEventResponse, error) {
	ev, err := s.getEventByShareCode(ctx, req.ShareCode)
	if err != nil {
		return nil, err
	}
	return &eventv1.GetEventResponse{Event: ev}, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, req *eventv1.UpdateEventRequest) (*eventv1.UpdateEventResponse, error) {
	var id, sc, title string
	var desc *string
	var eventDate *time.Time
	var createdAt time.Time
	var link *string

	err := s.db.Pool.QueryRow(ctx, `
		UPDATE events SET title = $1, description = $2, event_date = $3, link = $4
		WHERE id = $5
		RETURNING id, share_code, title, description, event_date, created_at, link`,
		req.Title, req.Description, timePtr(req.Date), req.Link, req.EventId,
	).Scan(&id, &sc, &title, &desc, &eventDate, &createdAt, &link)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("update event: %w", err)
	}

	authorName, _ := s.authorNameByEvent(ctx, id)

	return &eventv1.UpdateEventResponse{Event: &eventv1.Event{
		Id: id, ShareCode: sc, Title: title, Description: desc,
		Date: formatTimePtr(eventDate), AuthorName: authorName,
		CreatedAt: createdAt.Format(time.RFC3339), Link: link,
	}}, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, req *eventv1.DeleteEventRequest) (*eventv1.DeleteEventResponse, error) {
	tag, err := s.db.Pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, req.EventId)
	if err != nil {
		return nil, fmt.Errorf("delete event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return &eventv1.DeleteEventResponse{}, nil
}
func (s *EventService) ListMyEvents(ctx context.Context, userID string) ([]*eventv1.Event, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT e.id, e.share_code, e.title, e.description, e.event_date,
		       COALESCE(au.display_name, au.first_name),
		       e.created_at, e.link
		FROM events e
		JOIN participants mp ON mp.event_id = e.id
		LEFT JOIN participants ap ON ap.event_id = e.id AND ap.is_author = TRUE
		LEFT JOIN users au ON au.id = ap.user_id
		WHERE mp.user_id = $1
		ORDER BY e.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query my events: %w", err)
	}
	defer rows.Close()

	var events []*eventv1.Event
	for rows.Next() {
		var id, sc, title, authorName string
		var desc *string
		var eventDate *time.Time
		var createdAt time.Time
		var link *string

		if err := rows.Scan(&id, &sc, &title, &desc, &eventDate, &authorName, &createdAt, &link); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, &eventv1.Event{
			Id: id, ShareCode: sc, Title: title, Description: desc,
			Date: formatTimePtr(eventDate), AuthorName: authorName,
			CreatedAt: createdAt.Format(time.RFC3339), Link: link,
		})
	}
	return events, rows.Err()
}


func (s *EventService) GetEventByID(ctx context.Context, eventID string) (*eventv1.Event, error) {
	var id, sc, title, authorName string
	var desc *string
	var eventDate *time.Time
	var createdAt time.Time
	var link *string

	err := s.db.Pool.QueryRow(ctx, `
		SELECT e.id, e.share_code, e.title, e.description, e.event_date,
		       COALESCE(u.display_name, u.first_name),
		       e.created_at, e.link
		FROM events e
		LEFT JOIN participants ap ON ap.event_id = e.id AND ap.is_author = TRUE
		LEFT JOIN users u ON u.id = ap.user_id
		WHERE e.id = $1`, eventID,
	).Scan(&id, &sc, &title, &desc, &eventDate, &authorName, &createdAt, &link)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &eventv1.Event{
		Id: id, ShareCode: sc, Title: title, Description: desc,
		Date: formatTimePtr(eventDate), AuthorName: authorName,
		CreatedAt: createdAt.Format(time.RFC3339), Link: link,
	}, nil
}

func (s *EventService) getEventByShareCode(ctx context.Context, shareCode string) (*eventv1.Event, error) {
	var id, sc, title, authorName string
	var desc *string
	var eventDate *time.Time
	var createdAt time.Time
	var link *string

	err := s.db.Pool.QueryRow(ctx, `
		SELECT e.id, e.share_code, e.title, e.description, e.event_date,
		       COALESCE(u.display_name, u.first_name),
		       e.created_at, e.link
		FROM events e
		LEFT JOIN participants ap ON ap.event_id = e.id AND ap.is_author = TRUE
		LEFT JOIN users u ON u.id = ap.user_id
		WHERE e.share_code = $1`, shareCode,
	).Scan(&id, &sc, &title, &desc, &eventDate, &authorName, &createdAt, &link)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &eventv1.Event{
		Id: id, ShareCode: sc, Title: title, Description: desc,
		Date: formatTimePtr(eventDate), AuthorName: authorName,
		CreatedAt: createdAt.Format(time.RFC3339), Link: link,
	}, nil
}

var ErrNotFound = fmt.Errorf("not found")

func generateShareCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func timePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

// unused import guard — uuid may be needed later
var _ = uuid.Nil
