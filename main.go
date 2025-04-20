package main

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
)

var db *gorm.DB

func initDB() {
	dsn := "host=localhost user=postgres password=yourpassword dbname=PostgresAPI port=5432 sslmode=disable"
	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Task{}); err != nil {
		log.Fatalf("Could not migrate, %v", err)
	}
}

type Task struct {
	ID     string `gorm:"primaryKey" json:"id"`
	Task   string `json:"task"`
	IsDone bool   `json:"isDone"`
}

var Tasks = []Task{}

func GetTask(c echo.Context) error {
	return c.JSON(http.StatusOK, Tasks)
}

func PostTask(c echo.Context) error {
	var newTask Task
	if err := c.Bind(&newTask); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if newTask.Task == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "field 'Task' is required"})
	}

	newTask.ID = uuid.New().String()

	Tasks = append(Tasks, newTask)
	return c.JSON(http.StatusCreated, newTask)
}

func PatchTask(c echo.Context) error {
	id := c.Param("id")

	var UpdTask struct {
		Task   *string `json:"task"`
		IsDone *bool   `json:"isDone"`
	}
	if err := c.Bind(&UpdTask); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	for i := range Tasks {
		if Tasks[i].ID == id {
			if UpdTask.Task != nil {
				Tasks[i].Task = *UpdTask.Task
			}
			if UpdTask.IsDone != nil {
				Tasks[i].IsDone = *UpdTask.IsDone
			}
			return c.JSON(http.StatusOK, Tasks[i])
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
}

func DeleteTask(c echo.Context) error {
	id := c.Param("id")
	for i := range Tasks {
		if Tasks[i].ID == id {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			return c.NoContent(http.StatusNoContent)
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
}

func main() {

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	e.GET("api/tasks", GetTask)
	e.POST("api/tasks", PostTask)
	e.PATCH("api/tasks/:id", PatchTask)
	e.DELETE("api/tasks/:id", DeleteTask)

	if err := e.Start(":8080"); err != nil {
		log.Fatal("Error starting server", err)
	}
}
