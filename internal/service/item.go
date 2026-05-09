package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	itemv1 "github.com/atcdot/SharedPrep/api/item/v1"
	"github.com/atcdot/SharedPrep/internal/storage"

	"github.com/jackc/pgx/v5"
)

type ItemService struct {
	db *storage.Postgres
}

func NewItemService(db *storage.Postgres) *ItemService {
	return &ItemService{db: db}
}
func (s *ItemService) GetEventIDByItem(ctx context.Context, itemID string) (string, error) {
	var eventID string
	err := s.db.Pool.QueryRow(ctx, `SELECT event_id FROM items WHERE id = $1`, itemID).Scan(&eventID)
	if err != nil {
		return "", err
	}
	return eventID, nil
}


func scanItem(row pgx.Row, s *ItemService, ctx context.Context) (*itemv1.Item, error) {
	var id, evID, creatorID string
	var itemTitle string
	var qty int32
	var assignedID *string
	var createdAt time.Time

	err := row.Scan(&id, &evID, &creatorID, &itemTitle, &qty, &assignedID, &createdAt)
	if err != nil {
		return nil, err
	}

	it := &itemv1.Item{
		Id: id, EventId: evID, CreatorId: creatorID,
		CreatorName: resolveName(ctx, s.db, creatorID),
		Title: itemTitle, Quantity: qty,
		AssignedParticipantId: assignedID,
		CreatedAt: createdAt.Format(time.RFC3339),
	}
	if assignedID != nil {
		name := resolveName(ctx, s.db, *assignedID)
		it.AssignedParticipantName = &name
	}
	return it, nil
}

func (s *ItemService) CreateItem(ctx context.Context, participantID, eventID, title string, quantity int32) (*itemv1.Item, error) {
	row := s.db.Pool.QueryRow(ctx, `
		INSERT INTO items (event_id, creator_id, title, quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id, event_id, creator_id, title, quantity, assigned_participant_id, created_at`,
		eventID, participantID, title, quantity)

	it, err := scanItem(row, s, ctx)
	if err != nil {
		return nil, fmt.Errorf("insert item: %w", err)
	}
	return it, nil
}

func (s *ItemService) UpdateItem(ctx context.Context, itemID, title string, quantity int32) (*itemv1.Item, error) {
	row := s.db.Pool.QueryRow(ctx, `
		UPDATE items SET title = $1, quantity = $2
		WHERE id = $3
		RETURNING id, event_id, creator_id, title, quantity, assigned_participant_id, created_at`,
		title, quantity, itemID)

	it, err := scanItem(row, s, ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update item: %w", err)
	}
	return it, nil
}

func (s *ItemService) DeleteItem(ctx context.Context, itemID string) error {
	tag, err := s.db.Pool.Exec(ctx, `DELETE FROM items WHERE id = $1`, itemID)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *ItemService) ClaimItem(ctx context.Context, itemID, participantID string) (*itemv1.Item, error) {
	row := s.db.Pool.QueryRow(ctx, `
		UPDATE items SET assigned_participant_id = $1
		WHERE id = $2
		RETURNING id, event_id, creator_id, title, quantity, assigned_participant_id, created_at`,
		participantID, itemID)

	it, err := scanItem(row, s, ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("claim item: %w", err)
	}
	return it, nil
}

func (s *ItemService) UnclaimItem(ctx context.Context, itemID string) (*itemv1.Item, error) {
	row := s.db.Pool.QueryRow(ctx, `
		UPDATE items SET assigned_participant_id = NULL
		WHERE id = $1
		RETURNING id, event_id, creator_id, title, quantity, assigned_participant_id, created_at`,
		itemID)

	it, err := scanItem(row, s, ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("unclaim item: %w", err)
	}
	return it, nil
}

func (s *ItemService) ListItems(ctx context.Context, eventID string) ([]*itemv1.Item, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT i.id, i.event_id, i.creator_id,
		       COALESCE(cp_u.display_name, cp_u.first_name),
		       i.title, i.quantity, i.assigned_participant_id,
		       COALESCE(ap_u.display_name, ap_u.first_name),
		       i.created_at
		FROM items i
		JOIN participants cp ON cp.id = i.creator_id
		JOIN users cp_u ON cp_u.id = cp.user_id
		LEFT JOIN participants ap ON ap.id = i.assigned_participant_id
		LEFT JOIN users ap_u ON ap_u.id = ap.user_id
		WHERE i.event_id = $1
		ORDER BY i.created_at`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []*itemv1.Item
	for rows.Next() {
		var id, evID, creatorID, creatorName string
		var itemTitle string
		var qty int32
		var assignedID *string
		var assignedName *string
		var createdAt time.Time

		if err := rows.Scan(&id, &evID, &creatorID, &creatorName,
			&itemTitle, &qty, &assignedID, &assignedName, &createdAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		it := &itemv1.Item{
			Id: id, EventId: evID, CreatorId: creatorID, CreatorName: creatorName,
			Title: itemTitle, Quantity: qty,
			AssignedParticipantId: assignedID,
			CreatedAt: createdAt.Format(time.RFC3339),
		}
		if assignedID != nil && assignedName != nil {
			it.AssignedParticipantName = assignedName
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return items, nil
}

func resolveName(ctx context.Context, db *storage.Postgres, id string) string {
	if id == "" {
		return ""
	}
	var name string
	err := db.Pool.QueryRow(ctx, `
		SELECT COALESCE(u.display_name, u.first_name)
		FROM participants p JOIN users u ON u.id = p.user_id
		WHERE p.id = $1`, id).Scan(&name)
	if err != nil {
		return ""
	}
	return name
}
