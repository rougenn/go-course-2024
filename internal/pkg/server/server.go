package server

import (
	"encoding/json"
	"hw1/internal/pkg/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	storage *storage.Storage
	host    string
}

type Entry struct {
	Value string `json:"value"`
}

func New(host string, st *storage.Storage) *Server {
	s := &Server{
		host:    host,
		storage: st,
	}

	return s
}

func (r *Server) newAPI() *gin.Engine {
	engine := gin.New()

	engine.GET("/hello-world", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Hello world")
	})

	engine.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	engine.PUT("/scalar/set/:key", r.handlerSet)
	engine.GET("/scalar/get/:key", r.handlerGet)

	return engine
}

func (r *Server) handlerSet(ctx *gin.Context) {
	key := ctx.Param("key")

	var v Entry

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	r.storage.Set(key, v.Value)

	ctx.Status(http.StatusOK)
}

func (r *Server) handlerGet(ctx *gin.Context) {
	key := ctx.Param("key")

	v, err := r.storage.Get(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, Entry{
		Value: v,
	})
}

func (r *Server) Start() {
	err := r.newAPI().Run(r.host)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
