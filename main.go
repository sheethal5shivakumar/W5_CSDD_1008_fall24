package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Task represents the structure of a task
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending" or "completed"
}

// Global variables to hold tasks and manage concurrency
var (
	tasks  = []Task{} // Slice to store tasks
	nextID = 1        // ID counter for new tasks
	mu     sync.Mutex // Mutex for concurrent access
)

func main() {
	// Set up routes
	http.HandleFunc("/tasks", tasksHandler) // Handle tasks endpoint
	http.HandleFunc("/tasks/", taskHandler) // Handle individual task operations
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil) // Start the server
}

// tasksHandler handles the creation and listing of tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: // Get all tasks
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)

	case http.MethodPost: // Create a new task
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		task.ID = nextID            // Assign ID to new task
		nextID++                    // Increment ID for next task
		tasks = append(tasks, task) // Add task to slice
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task) // Return created task

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// taskHandler handles operations for a specific task identified by ID
func taskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/tasks/"):]
	id, err := strconv.Atoi(idStr) // Convert ID from string to integer
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	switch r.Method {
	case http.MethodGet: // Get a task by ID
		for _, task := range tasks {
			if task.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(task)
				return
			}
		}
		http.NotFound(w, r) // Task not found

	case http.MethodPut: // Update a task by ID
		var updatedTask Task
		if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for i, task := range tasks {
			if task.ID == id {
				tasks[i].Title = updatedTask.Title
				tasks[i].Description = updatedTask.Description
				tasks[i].Status = updatedTask.Status
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tasks[i]) // Return updated task
				return
			}
		}
		http.NotFound(w, r) // Task not found

	case http.MethodDelete: // Delete a task by ID
		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...) // Remove task from slice
				w.WriteHeader(http.StatusNoContent)       // No content response
				return
			}
		}
		http.NotFound(w, r) // Task not found

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
