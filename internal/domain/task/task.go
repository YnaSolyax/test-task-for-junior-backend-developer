package task

import "time"

type Status string

const (
	StatusNew              Status = "new"
	StatusInProgress       Status = "in_progress"
	StatusDone             Status = "done"
	StatusDeadlineExceeded Status = "deadline_exceeded"
)

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	ScheduledID *int64     `json:"scheduled_id,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone, StatusDeadlineExceeded:
		return true
	default:
		return false
	}
}

type ScheduledTask struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	CronExpression string    `json:"cron_expression"`
	NextRun        time.Time `json:"next_run"`
	CreatedAt      time.Time `json:"created_at"`
}
