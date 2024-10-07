package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "strconv"
)

type Task struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Status      string `json:"status"` // "pending" or "completed"
}

var (
    tasks  = []Task{}
    nextID = 1
    mu     sync.Mutex
)

func main() {
    http.HandleFunc("/tasks", tasksHandler)
    http.HandleFunc("/tasks/", taskHandler) // notice the trailing slash
    fmt.Println("Server started at :8080")
    http.ListenAndServe(":8080", nil)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(tasks)
    case http.MethodPost:
        var task Task
        if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        mu.Lock()
        task.ID = nextID
        nextID++
        tasks = append(tasks, task)
        mu.Unlock()
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(task)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/tasks/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    switch r.Method {
    case http.MethodGet:
        for _, task := range tasks {
            if task.ID == id {
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(task)
                return
            }
        }
        http.NotFound(w, r)

    case http.MethodPut:
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
                json.NewEncoder(w).Encode(tasks[i])
                return
            }
        }
        http.NotFound(w, r)

    case http.MethodDelete:
        for i, task := range tasks {
            if task.ID == id {
                tasks = append(tasks[:i], tasks[i+1:]...)
                w.WriteHeader(http.StatusNoContent)
                return
            }
        }
        http.NotFound(w, r)

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}