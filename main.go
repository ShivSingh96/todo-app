package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			completed BOOLEAN
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLFiles("index.html") // Make sure this path is correct

	// Serve the HTML page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API routes
	router.GET("/tasks", getTasks)
	router.POST("/tasks", createTask)
	router.PUT("/tasks/:id", updateTaskStatus)
	router.DELETE("/tasks/:id", deleteTask)

	router.Run(":8080")
}

func getTasks(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, completed FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var completed bool
		if err := rows.Scan(&id, &name, &completed); err != nil {
			log.Fatal(err)
		}
		task := map[string]interface{}{
			"id":        id,
			"name":      name,
			"completed": completed,
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT INTO tasks (name, completed) VALUES (?, ?)", task.Name, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": task.Name, "completed": false})
}

func updateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var task struct {
		Completed bool `json:"completed"`
	}
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE tasks SET completed = ? WHERE id = ?", task.Completed, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "completed": task.Completed})
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}
