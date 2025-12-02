package repository

import (
	"database/sql"
	"koda-shortlink-backend/internal/models"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.Session) error {
	query := `
		INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)
}

func (r *SessionRepository) FindByRefreshToken(token string) (*models.Session, error) {
	session := &models.Session{}
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1 AND expires_at > $2
	`

	err := r.db.QueryRow(query, token, time.Now()).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	return session, err
}

func (r *SessionRepository) DeleteByRefreshToken(token string) error {
	query := `DELETE FROM sessions WHERE refresh_token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *SessionRepository) DeleteExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	_, err := r.db.Exec(query, time.Now())
	return err
}
