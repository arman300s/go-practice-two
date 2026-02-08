package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"practice-one/internal/models"
	"practice-one/internal/store"
)

const (
	MaxTitleLength = 200
)

type TaskHandler struct {
	store *store.TaskStore
}

func NewTaskHandler(store *store.TaskStore) *TaskHandler {
	return &TaskHandler{store: store}
}

// GetTask handles GET /v1/tasks?id=X
// @Summary Get a single task
// @Description Get task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id query int true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /v1/tasks [get]
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// If no ID provided, return all tasks
		h.GetAllTasks(w, r)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid id"})
		return
	}

	task, err := h.store.GetByID(id)
	if err == store.ErrTaskNotFound {
		respondJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "task not found"})
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// GetAllTasks handles GET /v1/tasks or GET /v1/tasks?done=true
// @Summary Get all tasks
// @Description Get all tasks, optionally filtered by done status
// @Tags tasks
// @Accept json
// @Produce json
// @Param done query bool false "Filter by done status"
// @Success 200 {array} models.Task
// @Router /v1/tasks [get]
func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	doneParam := r.URL.Query().Get("done")

	var tasks []*models.Task

	if doneParam != "" {
		done, err := strconv.ParseBool(doneParam)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid done parameter"})
			return
		}
		tasks = h.store.GetByStatus(done)
	} else {
		tasks = h.store.GetAll()
	}

	respondJSON(w, http.StatusOK, tasks)
}

// CreateTask handles POST /v1/tasks
// @Summary Create a new task
// @Description Create a new task with title
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body models.CreateTaskRequest true "Task to create"
// @Success 201 {object} models.Task
// @Failure 400 {object} models.ErrorResponse
// @Router /v1/tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid title"})
		return
	}

	if len(req.Title) > MaxTitleLength {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("title exceeds maximum length of %d characters", MaxTitleLength),
		})
		return
	}

	task := h.store.Create(req.Title)
	respondJSON(w, http.StatusCreated, task)
}

// UpdateTask handles PATCH /v1/tasks?id=X
// @Summary Update a task
// @Description Update task's done status
// @Tags tasks
// @Accept json
// @Produce json
// @Param id query int true "Task ID"
// @Param task body models.UpdateTaskRequest true "Update data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /v1/tasks [patch]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "id parameter is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid id"})
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.store.Update(id, req.Done); err == store.ErrTaskNotFound {
		respondJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "task not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{Updated: true})
}

// DeleteTask handles DELETE /v1/tasks?id=X
// @Summary Delete a task
// @Description Delete task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id query int true "Task ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /v1/tasks [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "id parameter is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid id"})
		return
	}

	if err := h.store.Delete(id); err == store.ErrTaskNotFound {
		respondJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "task not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.SuccessResponse{Updated: true})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
