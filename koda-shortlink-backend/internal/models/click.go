package models

import (
	"time"
)

type Click struct {
	ID         int64     `json:"id" db:"id"`
	LinkID     int64     `json:"link_id" db:"link_id"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	Referer    *string   `json:"referer,omitempty" db:"referer"`
	Country    *string   `json:"country,omitempty" db:"country"`
	City       *string   `json:"city,omitempty" db:"city"`
	DeviceType *string   `json:"device_type,omitempty" db:"device_type"`
	Browser    *string   `json:"browser,omitempty" db:"browser"`
	OS         *string   `json:"os,omitempty" db:"os"`
	ClickedAt  time.Time `json:"clicked_at" db:"clicked_at"`
}

type ClickAnalytics struct {
	TotalClicks   int64                  `json:"total_clicks"`
	UniqueClicks  int64                  `json:"unique_clicks"`
	ClicksByDay   []ClicksByDay          `json:"clicks_by_day"`
	TopCountries  []CountryStats         `json:"top_countries"`
	TopCities     []CityStats            `json:"top_cities"`
	TopReferers   []RefererStats         `json:"top_referers"`
	DeviceStats   []DeviceStats          `json:"device_stats"`
	BrowserStats  []BrowserStats         `json:"browser_stats"`
	OSStats       []OSStats              `json:"os_stats"`
}

type ClicksByDay struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type CountryStats struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

type CityStats struct {
	City  string `json:"city"`
	Count int64  `json:"count"`
}

type RefererStats struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
}

type DeviceStats struct {
	DeviceType string `json:"device_type"`
	Count      int64  `json:"count"`
}

type BrowserStats struct {
	Browser string `json:"browser"`
	Count   int64  `json:"count"`
}

type OSStats struct {
	OS    string `json:"os"`
	Count int64  `json:"count"`
}
