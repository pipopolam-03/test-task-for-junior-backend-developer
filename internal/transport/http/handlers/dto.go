package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title         string                  `json:"title"`
	Description   string                  `json:"description"`
	Status        taskdomain.Status       `json:"status"`
	IntervalType  taskdomain.IntervalType `json:"interval_type"`
	IntervalDays  *int                    `json:"interval_days,omitempty"`
	DayOfMonth    *int                    `json:"day_of_month,omitempty"`
	IsEven        *bool                   `json:"is_even,omitempty"`
	SpecificDates []string                `json:"specific_dates,omitempty"`
}

type taskDTO struct {
	ID            int64                   `json:"id"`
	Title         string                  `json:"title"`
	Description   string                  `json:"description"`
	Status        taskdomain.Status       `json:"status"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	IntervalType  taskdomain.IntervalType `json:"interval_type"`
	IntervalDays  *int                    `json:"interval_days,omitempty"`
	DayOfMonth    *int                    `json:"day_of_month,omitempty"`
	IsEven        *bool                   `json:"is_even,omitempty"`
	SpecificDates []string                `json:"specific_dates,omitempty"`
	NextRunAt     *time.Time              `json:"next_run_at,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:            task.ID,
		Title:         task.Title,
		Description:   task.Description,
		Status:        task.Status,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
		IntervalType:  task.IntervalType,
		IntervalDays:  task.IntervalDays,
		DayOfMonth:    task.DayOfMonth,
		IsEven:        task.IsEven,
		SpecificDates: task.SpecificDates,
		NextRunAt:     task.NextRunAt,
	}
}
