package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (
			title,
			description,
			status,
			created_at,
			updated_at,
			interval_type,
			interval_days,
			day_of_month,
			is_even,
			specific_dates,
			next_run_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id, title, description, status, created_at, updated_at,
			interval_type, interval_days, day_of_month, is_even, specific_dates, next_run_at
	`

	specificDatesRaw, err := json.Marshal(task.SpecificDates)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.IntervalType,
		task.IntervalDays,
		task.DayOfMonth,
		task.IsEven,
		specificDatesRaw,
		task.NextRunAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT
			id, title, description, status, created_at, updated_at,
			interval_type, interval_days, day_of_month, is_even, specific_dates, next_run_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4,
			interval_type = $5,
			interval_days = $6,
			day_of_month = $7,
			is_even = $8,
			specific_dates = $9,
			next_run_at = $10
		WHERE id = $11
		RETURNING
			id, title, description, status, created_at, updated_at,
			interval_type, interval_days, day_of_month, is_even, specific_dates, next_run_at
	`

	specificDatesRaw, err := json.Marshal(task.SpecificDates)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.UpdatedAt,
		task.IntervalType,
		task.IntervalDays,
		task.DayOfMonth,
		task.IsEven,
		specificDatesRaw,
		task.NextRunAt,
		task.ID,
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT
			id, title, description, status, created_at, updated_at,
			interval_type, interval_days, day_of_month, is_even, specific_dates, next_run_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task             taskdomain.Task
		status           string
		intervalType     string
		specificDatesRaw []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&intervalType,
		&task.IntervalDays,
		&task.DayOfMonth,
		&task.IsEven,
		&specificDatesRaw,
		&task.NextRunAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.IntervalType = taskdomain.IntervalType(intervalType)

	if len(specificDatesRaw) > 0 {
		if err := json.Unmarshal(specificDatesRaw, &task.SpecificDates); err != nil {
			return nil, err
		}
	}

	if task.SpecificDates == nil {
		task.SpecificDates = []string{}
	}

	return &task, nil
}
