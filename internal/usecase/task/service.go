package task

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input, s.now())
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:         normalized.Title,
		Description:   normalized.Description,
		Status:        normalized.Status,
		IntervalType:  normalized.IntervalType,
		IntervalDays:  normalized.IntervalDays,
		DayOfMonth:    normalized.DayOfMonth,
		IsEven:        normalized.IsEven,
		SpecificDates: normalized.SpecificDates,
		NextRunAt:     normalized.NextRunAt,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input, s.now())
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:            id,
		Title:         normalized.Title,
		Description:   normalized.Description,
		Status:        normalized.Status,
		IntervalType:  normalized.IntervalType,
		IntervalDays:  normalized.IntervalDays,
		DayOfMonth:    normalized.DayOfMonth,
		IsEven:        normalized.IsEven,
		SpecificDates: normalized.SpecificDates,
		NextRunAt:     normalized.NextRunAt,
		UpdatedAt:     s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput, now time.Time) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	normalizedSchedule, err := normalizeSchedule(input.IntervalType, input.IntervalDays, input.DayOfMonth, input.IsEven, input.SpecificDates, now)
	if err != nil {
		return CreateInput{}, err
	}

	input.IntervalType = normalizedSchedule.IntervalType
	input.IntervalDays = normalizedSchedule.IntervalDays
	input.DayOfMonth = normalizedSchedule.DayOfMonth
	input.IsEven = normalizedSchedule.IsEven
	input.SpecificDates = normalizedSchedule.SpecificDates
	input.NextRunAt = normalizedSchedule.NextRunAt

	return input, nil
}

func validateUpdateInput(input UpdateInput, now time.Time) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	normalizedSchedule, err := normalizeSchedule(input.IntervalType, input.IntervalDays, input.DayOfMonth, input.IsEven, input.SpecificDates, now)
	if err != nil {
		return UpdateInput{}, err
	}

	input.IntervalType = normalizedSchedule.IntervalType
	input.IntervalDays = normalizedSchedule.IntervalDays
	input.DayOfMonth = normalizedSchedule.DayOfMonth
	input.IsEven = normalizedSchedule.IsEven
	input.SpecificDates = normalizedSchedule.SpecificDates
	input.NextRunAt = normalizedSchedule.NextRunAt

	return input, nil
}

type schedule struct {
	IntervalType  taskdomain.IntervalType
	IntervalDays  *int
	DayOfMonth    *int
	IsEven        *bool
	SpecificDates []string
	NextRunAt     *time.Time
}

func normalizeSchedule(intervalType taskdomain.IntervalType, intervalDays *int, dayOfMonth *int, isEven *bool, specificDates []string, now time.Time) (schedule, error) {
	if intervalType == "" {
		intervalType = taskdomain.IntervalNone
	}

	if !intervalType.Valid() {
		return schedule{}, fmt.Errorf("%w: invalid interval_type", ErrInvalidInput)
	}

	result := schedule{IntervalType: intervalType}

	// Обработка типов периодичности с проверкой ошибок
	switch intervalType {
	case taskdomain.IntervalNone:
		return result, nil
	case taskdomain.IntervalDaily:
		if intervalDays == nil || *intervalDays <= 0 {
			return schedule{}, fmt.Errorf("%w: interval_days must be >= 1 for daily interval", ErrInvalidInput)
		}
		days := *intervalDays
		result.IntervalDays = &days
		next := dayStartUTC(now).AddDate(0, 0, days)
		result.NextRunAt = &next
	case taskdomain.IntervalMonthly:
		if dayOfMonth == nil || *dayOfMonth < 1 || *dayOfMonth > 30 {
			return schedule{}, fmt.Errorf("%w: day_of_month must be from 1 to 30 for monthly interval", ErrInvalidInput)
		}
		day := *dayOfMonth
		result.DayOfMonth = &day
		next := nextMonthlyRun(now, day)
		result.NextRunAt = &next
	case taskdomain.IntervalParity:
		if isEven == nil {
			return schedule{}, fmt.Errorf("%w: is_even is required for parity interval", ErrInvalidInput)
		}
		parity := *isEven
		result.IsEven = &parity
		next := nextParityRun(now, parity)
		result.NextRunAt = &next
	case taskdomain.IntervalSpecific:
		if len(specificDates) == 0 {
			return schedule{}, fmt.Errorf("%w: specific_dates must not be empty for dates interval", ErrInvalidInput)
		}
		normalizedDates, nextRunAt, err := normalizeSpecificDates(specificDates, now)
		if err != nil {
			return schedule{}, err
		}
		result.SpecificDates = normalizedDates
		result.NextRunAt = nextRunAt
	}

	return result, nil
}

// Получение начала дня в UTC
func dayStartUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// Вычисление следующей даты для ежемесячного интервала
func nextMonthlyRun(now time.Time, day int) time.Time {
	current := dayStartUTC(now)
	year, month, _ := current.Date()
	candidate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if candidate.Before(current) {
		candidate = time.Date(year, month+1, day, 0, 0, 0, 0, time.UTC)
	}
	return candidate
}

// Вычисление следующей даты для интервала по четности
func nextParityRun(now time.Time, even bool) time.Time {
	candidate := dayStartUTC(now)
	for {
		if (candidate.Day()%2 == 0) == even {
			return candidate
		}
		candidate = candidate.AddDate(0, 0, 1)
	}
}

// Нормализация и валидация списка дат
func normalizeSpecificDates(raw []string, now time.Time) ([]string, *time.Time, error) {
	normalized := make([]string, 0, len(raw))
	parsed := make([]time.Time, 0, len(raw))
	seen := map[string]struct{}{}

	for _, dateStr := range raw {
		trimmed := strings.TrimSpace(dateStr)
		if trimmed == "" {
			continue
		}

		date, err := time.Parse("2006-01-02", trimmed)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: invalid date format in specific_dates (expected YYYY-MM-DD)", ErrInvalidInput)
		}

		formatted := date.UTC().Format("2006-01-02")
		if _, ok := seen[formatted]; ok {
			continue
		}
		seen[formatted] = struct{}{}
		normalized = append(normalized, formatted)
		parsed = append(parsed, dayStartUTC(date))
	}

	if len(normalized) == 0 {
		return nil, nil, fmt.Errorf("%w: specific_dates must contain at least one valid date", ErrInvalidInput)
	}

	slices.Sort(normalized)
	slices.SortFunc(parsed, func(a, b time.Time) int {
		if a.Before(b) {
			return -1
		}
		if a.After(b) {
			return 1
		}
		return 0
	})

	today := dayStartUTC(now)
	for _, d := range parsed {
		if !d.Before(today) {
			next := d
			return normalized, &next, nil
		}
	}

	return nil, nil, fmt.Errorf("%w: all specific_dates are in the past", ErrInvalidInput)
}
