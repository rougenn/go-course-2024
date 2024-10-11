package storage

import (
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type kind string

type val struct {
	k kind
	s string
	n int
}

const (
	KindString    = kind("S")
	KindInt       = kind("D")
	KindUndefined = kind("UNDEFINED")
)

type Storage struct {
	inner  map[string]*val
	logger *zap.Logger
}

func NewStorage() *Storage {
	// to disable logger while benchmarks
	logger, _ := zap.NewProduction(zap.IncreaseLevel(zapcore.DPanicLevel))
	// logger, _ := zap.NewProduction()

	// defer logger.Sync()

	logger.Info("new storage created")

	return &Storage{
		inner:  make(map[string]*val),
		logger: logger,
	}
}

func (r *Storage) Set(key, input_val string) {
	// defer r.logger.Sync()

	int_val, err := strconv.Atoi(input_val)
	if err == nil {
		r.inner[key] = &val{
			k: KindInt,
			n: int_val,
		}
		r.logger.Info("key obtained", zap.String("key", key),
			zap.Int("val", int_val), zap.String("type", string(KindInt)))

		return
	}
	r.inner[key] = &val{
		k: KindString,
		s: input_val,
	}
	r.logger.Info("key obtained", zap.String("key", key),
		zap.String("val", input_val),
		zap.String("type", string(KindString)))
}

func (r *Storage) GetValue(key string) (*val, bool) {
	// defer r.logger.Sync()

	val, ok := r.inner[key]
	if !ok {
		r.logger.Info("key value doesnt exists", zap.String("key", key))
		return nil, ok
	}
	if val.k == KindString {
		r.logger.Info("storage request", zap.String("key", key),
			zap.String("val", val.s), zap.String("type", string(val.k)))
	} else {
		r.logger.Info("storage request", zap.String("key", key),
			zap.Int("val", val.n), zap.String("type", string(val.k)))
	}

	return val, ok
}

func (r *Storage) Get(key string) *string {
	// defer r.logger.Sync()

	val, ok := r.GetValue(key)
	if !ok {
		return nil
	}
	switch val.k {
	case KindString:
		return &val.s
	case KindInt:
		output := strconv.Itoa(val.n)
		return &output
	default:
		return nil
	}
}

func (r *Storage) GetKind(key string) *kind {
	val, ok := r.GetValue(key)
	if !ok {
		return nil
	}
	return &val.k
}
