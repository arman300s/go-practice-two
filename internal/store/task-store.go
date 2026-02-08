package store

import (
	"errors"
	"sync"

	"practice-one/internal/models"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidID    = errors.New("invalid id")
)

type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[int]*models.Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]*models.Task),
		nextID: 1,
	}
}

func (s *TaskStore) Create(title string) *models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &models.Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[s.nextID] = task
	s.nextID++

	return task
}

func (s *TaskStore) GetByID(id int) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	taskCopy := *task
	return &taskCopy, nil
}

func (s *TaskStore) GetAll() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}

	return tasks
}

func (s *TaskStore) GetByStatus(done bool) []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.Task, 0)
	for _, task := range s.tasks {
		if task.Done == done {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}

	return tasks
}

func (s *TaskStore) Update(id int, done bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	task.Done = done
	return nil
}

func (s *TaskStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)
	return nil
}
