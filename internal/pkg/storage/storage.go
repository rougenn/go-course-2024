package storage

import (
	"errors"
	"slices"
	"sort"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type kind string

type val struct {
	value_type   kind
	string_value string
	int_value    int
}

const (
	KindString    = kind("S")
	KindInt       = kind("D")
	KindUndefined = kind("UNDEFINED")
)

type Storage struct {
	inner  map[string]*val
	arrays map[string][]int
	logger *zap.Logger
}

func NewStorage() *Storage {
	// to turn off the logger while benchmarks
	logger, _ := zap.NewProduction(zap.IncreaseLevel(zapcore.DPanicLevel))

	// logger, _ := zap.NewProduction()

	// defer logger.Sync()

	logger.Info("new storage created")

	return &Storage{
		inner:  make(map[string]*val),
		logger: logger,
		arrays: make(map[string][]int),
	}
}

func (r *Storage) Set(key, input_val string) {
	// defer r.logger.Sync()

	int_val, err := strconv.Atoi(input_val)
	if err == nil {
		r.inner[key] = &val{
			value_type: KindInt,
			int_value:  int_val,
		}
		r.logger.Info("key obtained", zap.String("key", key),
			zap.Int("val", int_val), zap.String("type", string(KindInt)))

		return
	}
	r.inner[key] = &val{
		value_type:   KindString,
		string_value: input_val,
	}
	r.logger.Info("key obtained", zap.String("key", key),
		zap.String("val", input_val),
		zap.String("type", string(KindString)))
}

var (
	ErrNoValue = errors.New("key value doesnt exist")
)

func (r *Storage) GetValue(key string) (*val, error) {
	// defer r.logger.Sync()

	val, ok := r.inner[key]
	if !ok {
		r.logger.Info("key value doesnt exist", zap.String("key", key))
		return nil, ErrNoValue
	}
	if val.value_type == KindString {
		r.logger.Info("storage request", zap.String("key", key),
			zap.String("val", val.string_value), zap.String("type", string(val.value_type)))
	} else {
		r.logger.Info("storage request", zap.String("key", key),
			zap.Int("val", val.int_value), zap.String("type", string(val.value_type)))
	}

	return val, nil
}

func (r *Storage) Get(key string) (string, error) {
	// defer r.logger.Sync()

	val, ok := r.GetValue(key)
	if ok != nil {
		return "", ok
	}
	switch val.value_type {
	case KindString:
		return val.string_value, nil
	case KindInt:
		output := strconv.Itoa(val.int_value)
		return output, nil
	default:
		return "", nil
	}
}

func (r *Storage) GetKind(key string) (kind, error) {
	val, err := r.GetValue(key)
	if err != nil {
		return "", err
	}
	return val.value_type, err
}

func (r *Storage) Rpush(key string, arr ...int) {

	r.arrays[key] = append(r.arrays[key], arr...)

	r.logger.Info("New elems added to RIGHT side of slice",
		zap.Int("count of elems", len(arr)), zap.String("key", key))
}

func (r *Storage) Lpush(key string, arr ...int) {

	r.arrays[key] = append(arr, r.arrays[key]...)

	r.logger.Info("New elems added to LEFT side of slice",
		zap.Int("count of elems", len(arr)), zap.String("key", key))
}

func (r *Storage) Raddtoset(key string, arr ...int) {
	to_append := []int{}
	sort.Slice(to_append, func(i, j int) bool {
		return i < j
	})

	prev := to_append[len(to_append)-1] + 1

	for _, elem := range arr {
		if prev == elem {
			continue
		}
		if ind := slices.Index(r.arrays[key], elem); ind == -1 {
			to_append = append(to_append, elem)
		}
		prev = elem
	}

	r.arrays[key] = append(r.arrays[key], to_append...)
}

func (r *Storage) Lpop(args ...int) {
	switch len(args) {
	case 1:

	}
}

func (r *Storage) Rpop(args ...int)
