package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/robfig/cron/v3"
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
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
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

func (s *Service) CreateScheduled(ctx context.Context, input CreateScheduledInput) (*taskdomain.ScheduledTask, error) {
	//hardcode в связи с нехваткой времени
	cronExpr := input.CronExpression
	if input.PeriodType != "" {
		switch input.PeriodType {
		case "daily":
			cronExpr = fmt.Sprintf("0 0 12 */%s * *", input.Value)
		case "monthly":
			cronExpr = fmt.Sprintf("0 0 12 %s * *", input.Value)
		case "even":
			cronExpr = "0 0 12 2-30/2 * *"
		case "odd":
			cronExpr = "0 0 12 1-31/2 * *"
		default:
			return nil, fmt.Errorf("unknown period type: %s", input.PeriodType)
		}
	}

	if cronExpr == "" {
		return nil, fmt.Errorf("either cron_expression or period_type with value is required")
	}

	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	model := &taskdomain.ScheduledTask{
		Title:          input.Title,
		Description:    input.Description,
		CronExpression: cronExpr,
		NextRun:        sched.Next(s.now()),
		CreatedAt:      s.now(),
	}

	return s.repo.CreateScheduled(ctx, model)
}

func (s *Service) SpawnTaskFromSchedule(ctx context.Context, scheduleID int64, title, desc string) error {
	_ = s.repo.MarkOverdueByScheduledID(ctx, scheduleID)
	model := &taskdomain.Task{
		Title:       title,
		Description: desc,
		Status:      taskdomain.StatusNew,
		ScheduledID: &scheduleID,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	_, err := s.repo.CreateWithSchedule(ctx, model)
	return err
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

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
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

func validateCreateInput(input CreateInput) (CreateInput, error) {
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

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
