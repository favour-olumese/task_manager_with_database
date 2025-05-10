package usecases

import (
	"context"
	domain "task_manager/Domain"
	repositories "task_manager/Repositories"

	"go.mongodb.org/mongo-driver/mongo"
)

type taskUsecase struct {
	taskRepo repositories.TaskRepository
}

// Create a new instance of TaskUsecase
func NewTaskUsecase(repo repositories.TaskRepository) domain.TaskUsecase {
	return &taskUsecase{
		taskRepo: repo,
	}
}

// Get all tasks.
func (repo *taskUsecase) GetAllTask(ctx context.Context) ([]domain.Task, error) {
	return repo.taskRepo.GetAllTask(ctx)
}

// Get specific task based on ID.
func (repo *taskUsecase) GetTaskByID(ctx context.Context, id string) (domain.Task, error) {
	return repo.taskRepo.GetTaskByID(ctx, id)
}

// Update existing task.
func (repo *taskUsecase) UpdateTask(ctx context.Context, id string, updatedTask domain.Task) error {
	return repo.taskRepo.UpdateTask(ctx, id, updatedTask)
}

// Delete a task
func (repo *taskUsecase) DeleteTask(ctx context.Context, id string) error {
	return repo.taskRepo.DeleteTask(ctx, id)
}

// Create new task.
func (repo *taskUsecase) NewTask(ctx context.Context, task domain.Task) (*mongo.InsertOneResult, error) {
	return repo.taskRepo.NewTask(ctx, task)
}
