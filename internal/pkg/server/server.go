package server

import (
	"hw1/internal/pkg/storage"
	"net/http"
	"strconv"

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

	engine.PUT("/scalar/set/:key", r.handlerSet)
	engine.GET("/scalar/get/:key", r.handlerGet)

	engine.PUT("/hset", r.handlerHset)
	engine.PUT("/rpush/:key", r.handlerRpush)
	engine.PUT("/lpush/:key", r.handlerLpush)
	engine.PUT("/raddtoset/:key", r.handlerRaddToSet)
	engine.DELETE("/deletesegment/:key", r.handlerDeleteSegment)
	engine.DELETE("/lpop/:key", r.handlerLpop)
	engine.DELETE("/rpop/:key", r.handlerRpop)
	engine.PUT("/lset/:key/:index", r.handlerLset)
	engine.GET("/lget/:key/:index", r.handlerLget)
	engine.PUT("/expire/:key/:seconds", r.handlerExpire)

	engine.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	return engine
}

func (r *Server) handlerSet(ctx *gin.Context) {
	key := ctx.Param("key")
	var v Entry

	if err := ctx.BindJSON(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := r.storage.Set(key, v.Value); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerGet(ctx *gin.Context) {
	key := ctx.Param("key")
	v, err := r.storage.Get(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, Entry{Value: v})
}

func (r *Server) handlerHset(ctx *gin.Context) {
	var data map[string]string
	if err := ctx.BindJSON(&data); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	args := []string{}
	for k, v := range data {
		args = append(args, k, v)
	}
	if err := r.storage.Hset(args...); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerRpush(ctx *gin.Context) {
	key := ctx.Param("key")
	var values []int
	if err := ctx.BindJSON(&values); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	r.storage.Rpush(key, values...)
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerLpush(ctx *gin.Context) {
	key := ctx.Param("key")
	var values []int
	if err := ctx.BindJSON(&values); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	r.storage.Lpush(key, values...)
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerRaddToSet(ctx *gin.Context) {
	key := ctx.Param("key")
	var values []int
	if err := ctx.BindJSON(&values); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := r.storage.Raddtoset(key, values...); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerDeleteSegment(ctx *gin.Context) {
	key := ctx.Param("key")
	left, _ := strconv.Atoi(ctx.Query("left"))
	right, _ := strconv.Atoi(ctx.Query("right"))

	_, err := r.storage.DeleteSegment(key, left, right)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerLpop(ctx *gin.Context) {
	key := ctx.Param("key")
	result, err := r.storage.Lpop(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (r *Server) handlerRpop(ctx *gin.Context) {
	key := ctx.Param("key")
	result, err := r.storage.Rpop(key)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (r *Server) handlerLset(ctx *gin.Context) {
	key := ctx.Param("key")
	index, _ := strconv.Atoi(ctx.Param("index"))
	var value int
	if err := ctx.BindJSON(&value); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := r.storage.Lset(key, index, value); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (r *Server) handlerLget(ctx *gin.Context) {
	key := ctx.Param("key")
	index, _ := strconv.Atoi(ctx.Param("index"))

	value, err := r.storage.Lget(key, index)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, value)
}

func (r *Server) handlerExpire(ctx *gin.Context) {
	key := ctx.Param("key")
	seconds, _ := strconv.ParseInt(ctx.Param("seconds"), 10, 64)

	if !r.storage.Expire(key, seconds) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (r *Server) Start() {
	r.newAPI().Run(r.host)
}
