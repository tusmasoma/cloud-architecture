package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/tusmasoma/cloud-architecture/spa-infrastructure/api/config"
	"github.com/tusmasoma/go-tech-dojo/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found", log.Ferror(err))
	}

	var addr string
	flag.StringVar(&addr, "addr", ":8080", "tcp host:port to connect")
	flag.Parse()

	_, cancelMain := context.WithCancel(context.Background())
	defer cancelMain()

	h := newTodoHandler()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origin"},
		ExposeHeaders:    []string{"Link", "Authorization"},
		AllowCredentials: true,
		MaxAge:           time.Duration(300) * time.Second,
	}))

	r.GET("/api/todos", h.listTodos)
	r.POST("/api/todos", h.createTodo)
	r.PATCH("/api/todos/:id", h.updateTodo)
	r.DELETE("/api/todos/:id", h.deleteTodo)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Info("Server running...")

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server failed", log.Ferror(err))
			return
		}
	}()

	<-signalCtx.Done()
	log.Info("Server stopping...")

	tctx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(tctx); err != nil {
		log.Error("Failed to shutdown http server", log.Ferror(err))
	}
	log.Info("Server exited")
}

type Todo struct {
	ID        string `json:"_id" gorm:"primaryKey"`
	Completed bool   `json:"completed" gorm:"column:completed"` // タスクが完了したかどうか
	Body      string `json:"body" gorm:"column:body"`           // タスクの内容(text)
}

func NewTodo(completed bool, body string) *Todo {
	return &Todo{
		ID:        uuid.New().String(),
		Completed: completed,
		Body:      body,
	}
}

type todoHandler struct {
	db *gorm.DB
}

func newTodoHandler() *todoHandler {
	ctx := context.Background()
	conf, err := config.NewDBConfig(ctx)
	if err != nil {
		log.Error("Failed to load database config", log.Ferror(err))
		return nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		conf.User, conf.Password, conf.Host, conf.Port, conf.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{}) // ping is automatically called
	if err != nil {
		return nil
	}

	if err := db.AutoMigrate(&Todo{}); err != nil { // migrate the schema
		log.Error("Failed to migrate database", log.Ferror(err))
		return nil
	}

	return &todoHandler{db: db}
}

func (h *todoHandler) listTodos(c *gin.Context) {
	ctx := c.Request.Context()

	var todos []Todo
	if err := h.db.WithContext(ctx).Find(&todos).Error; err != nil {
		log.Error("Failed to list todos", log.Ferror(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

type createTodoRequest struct {
	Body      string `json:"body"`
	Completed bool   `json:"completed"`
}

func (h *todoHandler) createTodo(c *gin.Context) {
	ctx := c.Request.Context()

	var req createTodoRequest
	if err := c.BindJSON(&req); err != nil {
		log.Error("Failed to bind request", log.Ferror(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}

	todo := NewTodo(req.Completed, req.Body)
	if err := h.db.WithContext(ctx).Create(todo).Error; err != nil {
		log.Error("Failed to create todo", log.Ferror(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

func (h *todoHandler) updateTodo(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var todo Todo
	if err := h.db.WithContext(ctx).First(&todo, "id = ?", id).Error; err != nil {
		log.Error("Failed to find todo", log.Ferror(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo.Completed = !todo.Completed

	if err := h.db.WithContext(ctx).Save(&todo).Error; err != nil {
		log.Error("Failed to update todo", log.Ferror(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})

}

func (h *todoHandler) deleteTodo(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.db.WithContext(ctx).Delete(&Todo{}, "id = ?", id).Error; err != nil {
		log.Error("Failed to delete todo", log.Ferror(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
