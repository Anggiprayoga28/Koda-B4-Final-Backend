package repository

import (
	"database/sql"
	"errors"
	"koda-shortlink-backend/internal/models"
	"time"
)

type ShortLinkRepository struct {
	db *sql.DB
}

func NewShortLinkRepository(db *sql.DB) *ShortLinkRepository {
	return &ShortLinkRepository{db: db}
}

func (r *ShortLinkRepository) Create(link *models.ShortLink) error {
	query := `
		INSERT INTO short_links (short_code, destination, user_id, title, description, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at, click_count
	`

	err := r.db.QueryRow(
		query,
		link.ShortCode,
		link.Destination,
		link.UserID,
		link.Title,
		link.Description,
		link.IsActive,
		link.ExpiresAt,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt, &link.ClickCount)

	return err
}

func (r *ShortLinkRepository) FindByShortCode(shortCode string) (*models.ShortLink, error) {
	link := &models.ShortLink{}
	query := `
		SELECT id, short_code, destination, user_id, title, description, is_active, 
		       click_count, created_at, updated_at, expires_at
		FROM short_links
		WHERE short_code = $1
	`

	err := r.db.QueryRow(query, shortCode).Scan(
		&link.ID,
		&link.ShortCode,
		&link.Destination,
		&link.UserID,
		&link.Title,
		&link.Description,
		&link.IsActive,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
		&link.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("short link not found")
	}

	return link, err
}

func (r *ShortLinkRepository) FindByID(id int64) (*models.ShortLink, error) {
	link := &models.ShortLink{}
	query := `
		SELECT id, short_code, destination, user_id, title, description, is_active, 
		       click_count, created_at, updated_at, expires_at
		FROM short_links
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&link.ID,
		&link.ShortCode,
		&link.Destination,
		&link.UserID,
		&link.Title,
		&link.Description,
		&link.IsActive,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
		&link.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("short link not found")
	}

	return link, err
}

func (r *ShortLinkRepository) FindByUser(userID int64, page, pageSize int) ([]models.ShortLink, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	countQuery := `SELECT COUNT(*) FROM short_links WHERE user_id = $1`
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, short_code, destination, user_id, title, description, is_active, 
		       click_count, created_at, updated_at, expires_at
		FROM short_links
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var links []models.ShortLink
	for rows.Next() {
		var link models.ShortLink
		err := rows.Scan(
			&link.ID,
			&link.ShortCode,
			&link.Destination,
			&link.UserID,
			&link.Title,
			&link.Description,
			&link.IsActive,
			&link.ClickCount,
			&link.CreatedAt,
			&link.UpdatedAt,
			&link.ExpiresAt,
		)
		if err != nil {
			return nil, 0, err
		}
		links = append(links, link)
	}

	return links, total, nil
}

func (r *ShortLinkRepository) Update(link *models.ShortLink) error {
	query := `
		UPDATE short_links
		SET destination = $1, title = $2, description = $3, is_active = $4, 
		    expires_at = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`

	result, err := r.db.Exec(
		query,
		link.Destination,
		link.Title,
		link.Description,
		link.IsActive,
		link.ExpiresAt,
		link.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("short link not found")
	}

	return nil
}

func (r *ShortLinkRepository) Delete(id int64, userID int64) error {
	query := `DELETE FROM short_links WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("short link not found or unauthorized")
	}

	return nil
}

func (r *ShortLinkRepository) ShortCodeExists(shortCode string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code = $1)`
	err := r.db.QueryRow(query, shortCode).Scan(&exists)
	return exists, err
}

func (r *ShortLinkRepository) IncrementClickCount(linkID int64) error {
	query := `UPDATE short_links SET click_count = click_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, linkID)
	return err
}

func (r *ShortLinkRepository) GetDashboardStats(userID int64) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := r.db.QueryRow(`SELECT COUNT(*) FROM short_links WHERE user_id = $1`, userID).Scan(&stats.TotalLinks)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT COALESCE(SUM(sl.click_count), 0)
		FROM short_links sl
		WHERE sl.user_id = $1
	`
	err = r.db.QueryRow(query, userID).Scan(&stats.TotalVisits)
	if err != nil {
		return nil, err
	}

	if stats.TotalLinks > 0 {
		stats.AvgClickRate = float64(stats.TotalVisits) / float64(stats.TotalLinks)
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	visitQuery := `
		SELECT DATE(c.clicked_at) as date, COUNT(*) as visits
		FROM clicks c
		JOIN short_links sl ON c.link_id = sl.id
		WHERE sl.user_id = $1 AND c.clicked_at >= $2
		GROUP BY DATE(c.clicked_at)
		ORDER BY date
	`

	rows, err := r.db.Query(visitQuery, userID, sevenDaysAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chart models.DailyVisitChart
		err := rows.Scan(&chart.Date, &chart.Visits)
		if err != nil {
			return nil, err
		}
		stats.Last7DaysVisits = append(stats.Last7DaysVisits, chart)
	}

	twoWeeksAgo := time.Now().AddDate(0, 0, -14)
	var lastWeekVisits, thisWeekVisits int64

	r.db.QueryRow(`
		SELECT COUNT(*) FROM clicks c
		JOIN short_links sl ON c.link_id = sl.id
		WHERE sl.user_id = $1 AND c.clicked_at >= $2 AND c.clicked_at < $3
	`, userID, twoWeeksAgo, sevenDaysAgo).Scan(&lastWeekVisits)

	r.db.QueryRow(`
		SELECT COUNT(*) FROM clicks c
		JOIN short_links sl ON c.link_id = sl.id
		WHERE sl.user_id = $1 AND c.clicked_at >= $2
	`, userID, sevenDaysAgo).Scan(&thisWeekVisits)

	if lastWeekVisits > 0 {
		stats.VisitsGrowth = ((float64(thisWeekVisits) - float64(lastWeekVisits)) / float64(lastWeekVisits)) * 100
	}

	return stats, nil
}
