package storage

import (
	"go.uber.org/zap"
	"strconv"
)

type val struct {
	kind string
	s    string
	n    int
}

type Storage struct {
	inner  map[string]*val
	logger *zap.Logger
}

func NewStorage() *Storage {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("new storage created")

	return &Storage{
		inner:  make(map[string]*val),
		logger: logger,
	}
}

func (r *Storage) Set(key, input_val string) {
	defer r.logger.Sync()

	int_val, err := strconv.Atoi(input_val)
	if err == nil {
		r.inner[key] = &val{
			kind: "int",
			n:    int_val,
		}
		r.logger.Info("key obtained", zap.String("key", key), zap.Int("val", int_val), zap.String("type", "int"))
		return
	}
	r.inner[key] = &val{
		kind: "string",
		s:    input_val,
	}
	r.logger.Info("key obtained", zap.String("key", key), zap.String("val", input_val), zap.String("type", "string"))
}

func (r *Storage) GetValue(key string) (*val, bool) {
	defer r.logger.Sync()

	val, ok := r.inner[key]
	if !ok {
		r.logger.Info("key value doesnt exists", zap.String("key", key))
		return nil, ok
	}
	if val.kind == "string" {
		r.logger.Info("storage request", zap.String("key", key), zap.String("val", val.s), zap.String("type", val.kind))
	} else {
		r.logger.Info("storage request", zap.String("key", key), zap.Int("val", val.n), zap.String("type", val.kind))
	}

	return val, ok
}

func (r *Storage) Get(key string) *string {
	defer r.logger.Sync()

	val, ok := r.GetValue(key)
	if !ok {
		return nil
	}
	if val.kind == "string" {
		return &val.s
	}
	output := strconv.Itoa(val.n)
	return &output
}

func (r *Storage) GetKind(key string) *string {
	val, ok := r.GetValue(key)
	if !ok {
		return nil
	}
	return &val.kind
}
