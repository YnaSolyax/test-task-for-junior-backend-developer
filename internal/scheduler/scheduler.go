package scheduler

import (
	"context"
	"log/slog"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron    *cron.Cron
	usecase taskusecase.Usecase
	logger  *slog.Logger
}

func NewScheduler(usecase taskusecase.Usecase, logger *slog.Logger) *Scheduler {
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:    c,
		usecase: usecase,
		logger:  logger,
	}
}

func (s *Scheduler) Start() {
	s.logger.Info("Starting cron scheduler...")
	s.cron.Start()
}

func (s *Scheduler) AddTask(ctx context.Context, st taskdomain.ScheduledTask) error {
	_, err := s.cron.AddFunc(st.CronExpression, func() {
		s.logger.Debug("Cron triggered", "scheduled_id", st.ID, "title", st.Title)

		err := s.usecase.SpawnTaskFromSchedule(context.Background(), st.ID, st.Title, st.Description)
		if err != nil {
			s.logger.Error("Failed to spawn task from schedule", "err", err, "scheduled_id", st.ID)
		}
	})
	return err
}
