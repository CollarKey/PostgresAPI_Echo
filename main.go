package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

var db *gorm.DB

func initDB() {
	dsn := "host=localhost user=postgres password=yourpassword dbname=tasks port=5432 sslmode=disable"
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
	ID     uint   `gorm:"primaryKey" json:"id"`
	Task   string `json:"task"`
	IsDone bool   `json:"isDone"`
}

func ParseID(idParam string) (uint, error) {
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func GetTask(c echo.Context) error {
	var task []Task

	if err := db.Find(&task).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not get task"})
	}
	return c.JSON(http.StatusOK, task)
}

func PostTask(c echo.Context) error {
	var newTask Task
	if err := c.Bind(&newTask); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if newTask.Task == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "field 'Task' is required"})
	}

	if err := db.Create(&newTask).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create task"})
	}
	return c.JSON(http.StatusCreated, newTask)
}

func UpdateTask(c echo.Context) error {
	id, err := ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	var task Task
	if err := db.First(&task, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Could not find task"})
	}

	var UpdTask struct {
		Task   *string `json:"task"`
		IsDone *bool   `json:"isDone"`
	}
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if UpdTask.Task != nil {
		task.Task = *UpdTask.Task
	}
	if UpdTask.IsDone != nil {
		task.IsDone = *UpdTask.IsDone
	}

	if err := db.Save(&task).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not update task"})
	}

	return c.JSON(http.StatusOK, task)
}

func DeleteTask(c echo.Context) error {
	id, err := ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	result := db.Delete(&Task{}, "id = ?", id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not delete task"})
	}

	return c.NoContent(http.StatusNoContent)
}

func main() {
	initDB()

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	e.GET("api/tasks", GetTask)
	e.POST("api/tasks", PostTask)
	e.PATCH("api/tasks/:id", UpdateTask)
	e.DELETE("api/tasks/:id", DeleteTask)

	if err := e.Start(":8080"); err != nil {
		log.Fatal("Error starting server", err)
	}
}
