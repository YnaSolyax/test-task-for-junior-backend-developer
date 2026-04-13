package postgres

import (
	"context"
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
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, status, scheduled_id, deadline, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, scheduled_id, deadline, created_at, updated_at
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
        	updated_at = $4
    	WHERE id = $5
    	RETURNING id, title, description, status, scheduled_id, deadline, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
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
		SELECT id, title, description, status, scheduled_id, deadline, created_at, updated_at
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
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.ScheduledID,
		&task.Deadline,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	return &task, nil
}

func (r *Repository) CreateScheduled(ctx context.Context, st *taskdomain.ScheduledTask) (*taskdomain.ScheduledTask, error) {
	const query = `
        INSERT INTO scheduled_tasks (title, description, cron_expression, next_run, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, title, description, cron_expression, next_run, created_at
    `
	row := r.pool.QueryRow(ctx, query, st.Title, st.Description, st.CronExpression, st.NextRun, st.CreatedAt)
	return scanScheduledTask(row)
}

func (r *Repository) CreateWithSchedule(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
        INSERT INTO tasks (title, description, status, scheduled_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, title, description, status, scheduled_id, deadline, created_at, updated_at
    `

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.ScheduledID, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}
	return created, nil
}

func (r *Repository) ListScheduled(ctx context.Context) ([]taskdomain.ScheduledTask, error) {
	const query = `SELECT id, title, description, cron_expression, next_run, created_at FROM scheduled_tasks`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.ScheduledTask
	for rows.Next() {
		task, err := scanScheduledTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	return tasks, nil
}

func scanScheduledTask(scanner taskScanner) (*taskdomain.ScheduledTask, error) {
	var task taskdomain.ScheduledTask
	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.CronExpression,
		&task.NextRun,
		&task.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *Repository) MarkOverdueByScheduledID(ctx context.Context, scheduledID int64) error {
	const query = `
        UPDATE tasks 
        SET status = 'deadline_exceeded', updated_at = NOW() 
        WHERE scheduled_id = $1 AND status IN ('new', 'in_progress')
    `
	_, err := r.pool.Exec(ctx, query, scheduledID)
	return err
}
