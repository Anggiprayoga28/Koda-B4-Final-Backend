package repository

import (
	"database/sql"
	"koda-shortlink-backend/internal/models"
)

type ClickRepository struct {
	db *sql.DB
}

func NewClickRepository(db *sql.DB) *ClickRepository {
	return &ClickRepository{db: db}
}

func (r *ClickRepository) Create(click *models.Click) error {
	query := `
		INSERT INTO clicks (link_id, ip_address, user_agent, referer, country, city, device_type, browser, os)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, clicked_at
	`

	return r.db.QueryRow(
		query,
		click.LinkID,
		click.IPAddress,
		click.UserAgent,
		click.Referer,
		click.Country,
		click.City,
		click.DeviceType,
		click.Browser,
		click.OS,
	).Scan(&click.ID, &click.ClickedAt)
}
