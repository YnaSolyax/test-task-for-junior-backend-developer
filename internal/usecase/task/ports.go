package task

import (
	"context"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	CreateScheduled(ctx context.Context, st *taskdomain.ScheduledTask) (*taskdomain.ScheduledTask, error)
	ListScheduled(ctx context.Context) ([]taskdomain.ScheduledTask, error)
	CreateWithSchedule(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	MarkOverdueByScheduledID(ctx context.Context, scheduledID int64) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	CreateScheduled(ctx context.Context, input CreateScheduledInput) (*taskdomain.ScheduledTask, error)
	SpawnTaskFromSchedule(ctx context.Context, scheduleID int64, title, desc string) error
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type CreateScheduledInput struct {
	Title          string
	Description    string
	CronExpression string
	PeriodType     string
	Value          string
}
