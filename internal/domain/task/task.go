package task

import "time"

type Status string

type IntervalType string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

const (
	IntervalNone     IntervalType = "none"
	IntervalDaily    IntervalType = "daily"
	IntervalMonthly  IntervalType = "monthly"
	IntervalParity   IntervalType = "parity"
	IntervalSpecific IntervalType = "dates"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	IntervalType  IntervalType `json:"interval_type"`
	IntervalDays  *int         `json:"interval_days,omitempty"`
	DayOfMonth    *int         `json:"day_of_month,omitempty"`
	IsEven        *bool        `json:"is_even,omitempty"`
	SpecificDates []string     `json:"specific_dates,omitempty"`
	NextRunAt     *time.Time   `json:"next_run_at,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t IntervalType) Valid() bool {
	switch t {
	case IntervalNone, IntervalDaily, IntervalMonthly, IntervalParity, IntervalSpecific:
		return true
	default:
		return false
	}
}
