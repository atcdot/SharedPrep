package service

import (
	"context"

	"github.com/atcdot/SharedPrep/internal/storage"
)

type UserService struct {
	db *storage.Postgres
}

func NewUserService(db *storage.Postgres) *UserService {
	return &UserService{db: db}
}

type User struct {
	ID         string
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	PhotoURL   string
}

func (s *UserService) FindByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	u := &User{}
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, photo_url
		FROM users WHERE telegram_id = $1
	`, telegramID).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.PhotoURL)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, photo_url
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.PhotoURL)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) FindParticipantByUserAndEvent(ctx context.Context, userID, eventID string) (string, error) {
	var id string
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id FROM participants WHERE user_id = $1 AND event_id = $2
	`, userID, eventID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}
