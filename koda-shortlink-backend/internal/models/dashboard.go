package models

type DashboardStats struct {
	TotalLinks      int64              `json:"total_links"`
	TotalVisits     int64              `json:"total_visits"`
	AvgClickRate    float64            `json:"avg_click_rate"`
	VisitsGrowth    float64            `json:"visits_growth"`
	Last7DaysVisits []DailyVisitChart  `json:"last_7_days_visits"`
}

type DailyVisitChart struct {
	Date  string `json:"date"`
	Visits int64 `json:"visits"`
}
