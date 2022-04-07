package restful

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/comet"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"

	"github.com/gin-gonic/gin"
	ginlog "github.com/quanxiang-cloud/cabin/tailormade/gin"
)

const (
	// DebugMode indicates mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates mode is release.
	ReleaseMode = "release"
)

// Router Router
type Router struct {
	c      *config.Config
	engine *gin.Engine
}

// NewRouter newRouter
func NewRouter(ctx context.Context, c *config.Config) (*Router, error) {

	engine, err := newRouter(c)
	if err != nil {
		return nil, err
	}
	v1 := engine.Group("/api/v1/entrepot/")

	factor, err := comet.NewFactor(ctx, c)
	if err != nil {
		return nil, err
	}
	batchTask, err := NewBatchTask(c, factor.Factor)
	if err != nil {
		return nil, err
	}
	task := v1.Group("/task")
	task.POST("/create/:command", batchTask.CreatTask)
	task.POST("/list", batchTask.GetList)
	task.POST("/get/:taskID", batchTask.GetByID)
	task.POST("/subscribe", batchTask.Subscribe)
	task.POST("/processing", batchTask.GetProcessing)
	task.POST("/delete/:taskID", batchTask.Delete)

	return &Router{
		c:      c,
		engine: engine,
	}, nil

}

func newRouter(c *config.Config) (*gin.Engine, error) {
	if c.Model == "" || (c.Model != ReleaseMode && c.Model != DebugMode) {
		c.Model = ReleaseMode
	}
	gin.SetMode(c.Model)
	engine := gin.New()
	engine.Use(ginlog.LoggerFunc(), ginlog.LoggerFunc())
	return engine, nil
}

// Run Start the service
func (r *Router) Run() {
	r.engine.Run(r.c.Port)
}

// Close Close the service
func (r *Router) Close() {

}
