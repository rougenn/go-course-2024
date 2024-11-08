package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	inner  map[string]*val  `json:"inner"`
	arrays map[string][]int `json:"arrays"`
	logger *zap.Logger      `json:"-"`
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
	ErrKeyDoesntExist  = errors.New("key value doesnt exist")
	ErrIndexOutOfRange = errors.New("index is out of range")
)

func (r *Storage) GetValue(key string) (*val, error) {
	// defer r.logger.Sync()

	val, ok := r.inner[key]
	if !ok {
		r.logger.Info("key value doesnt exist", zap.String("key", key))
		return nil, ErrKeyDoesntExist
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

	for _, elem := range arr {
		exists := false
		for _, i := range r.arrays[key] {
			if elem == i {
				exists = true
				break
			}
		}
		if !exists {
			r.arrays[key] = append(r.arrays[key], elem)
		}
	}
	r.logger.Info("New elements added", zap.String("key", key))
}

func (r *Storage) DeleteSegment(key string, l int, ri int) ([]int, error) {
	leng := len(r.arrays[key])

	if leng == 0 {
		return []int{}, nil
	}

	if ri < 0 {
		ri = leng + ri%leng
	}

	// fmt.Println(l, ri, r.arrays[key])

	if l > ri || l >= leng || ri >= leng {
		r.logger.Error("invalid indexes")
		return []int{}, ErrIndexOutOfRange
	}

	left_part := r.arrays[key][:l]
	right_part := r.arrays[key][ri+1:]

	deleted := make([]int, ri-l+1)
	copy(deleted, r.arrays[key][l:ri+1])
	// fmt.Println(left_part, deleted, right_part)

	r.arrays[key] = append(left_part, right_part...)
	r.logger.Info("Some elems has deleted from array",
		zap.String("key", key), zap.Int("left index", l),
		zap.Int("right index", ri))
	return deleted, nil
}

func (r *Storage) Lpop(key string, args ...int) ([]int, error) {
	_, ok := r.arrays[key]
	if !ok {
		return []int{}, ErrKeyDoesntExist
	}

	length := len(r.arrays[key])
	switch le := len(args); le {
	case 0:
		cnt := 1

		if cnt > len(r.arrays[key]) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := r.arrays[key][:cnt]

		r.arrays[key] = r.arrays[key][cnt:]
		r.logger.Info("deleted elems from left",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 1:
		cnt := args[0]

		if cnt > len(r.arrays[key]) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := make([]int, cnt)
		copy(deleted, r.arrays[key][:cnt])

		r.arrays[key] = r.arrays[key][cnt:]
		r.logger.Info("deleted elems from left",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 2:
		return r.DeleteSegment(key, args[0], args[1])

	default:
		r.logger.Error("Invalid count of arguments, max count is 3")
		return []int{}, ErrIndexOutOfRange
	}
}

func (r *Storage) Rpop(key string, args ...int) ([]int, error) {
	_, ok := r.arrays[key]
	if !ok {
		return []int{}, ErrKeyDoesntExist
	}
	length := len(r.arrays[key])
	switch le := len(args); le {
	case 0:
		cnt := 1

		if cnt > len(r.arrays[key]) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := make([]int, cnt)
		// fmt.Println(length, cnt, r.arrays[key])
		copy(deleted, r.arrays[key][length-cnt:length])

		r.arrays[key] = r.arrays[key][:length-cnt]
		r.logger.Info("deleted elems from right",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 1:

		cnt := args[0]

		if cnt > len(r.arrays[key]) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := make([]int, cnt)
		// fmt.Println(length, cnt, r.arrays[key])
		copy(deleted, r.arrays[key][length-cnt:length])

		// 0 1 2 3 4
		// 3
		// 0 1 2
		// [0, length - cnt + 1)
		r.arrays[key] = r.arrays[key][:length-cnt]
		r.logger.Info("deleted elems from right",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 2:
		return r.DeleteSegment(key, args[0], args[1])

	default:
		r.logger.Error("Invalid count of arguments, max count is 3")
		return []int{}, ErrIndexOutOfRange
	}
}

func (r *Storage) Lset(key string, index int, new_val int) error {
	arr, ok := r.arrays[key]
	if !ok {
		r.logger.Error(ErrKeyDoesntExist.Error())
		return ErrKeyDoesntExist
	}
	if index >= len(arr) {
		r.logger.Error(ErrIndexOutOfRange.Error())
		return ErrIndexOutOfRange
	}
	arr[index] = new_val
	r.logger.Info("element changed", zap.String("key", key),
		zap.Int("index", index), zap.Int("new val", new_val))
	return nil
}

func (r *Storage) Lget(key string, index int) (int, error) {
	arr, ok := r.arrays[key]
	if !ok {
		r.logger.Error(ErrKeyDoesntExist.Error())
		return 0, ErrKeyDoesntExist
	}
	if index >= len(arr) {
		r.logger.Error(ErrIndexOutOfRange.Error())
		return 0, ErrIndexOutOfRange
	}
	r.logger.Info("value requested", zap.String("key", key),
		zap.Int("index", index))
	return arr[index], nil
}

func (v *val) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ValueType   kind   `json:"value_type"`
		StringValue string `json:"string_value,omitempty"`
		IntValue    int    `json:"int_value,omitempty"`
	}{
		ValueType:   v.value_type,
		StringValue: v.string_value,
		IntValue:    v.int_value,
	})
}

func (v *val) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ValueType   kind   `json:"value_type"`
		StringValue string `json:"string_value"`
		IntValue    int    `json:"int_value"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	v.value_type = aux.ValueType
	v.string_value = aux.StringValue
	v.int_value = aux.IntValue

	return nil
}

func (r *Storage) MarshalJSON() ([]byte, error) {
	type storageAlias Storage
	return json.Marshal(&struct {
		Inner  map[string]*val  `json:"inner"`
		Arrays map[string][]int `json:"arrays"`
		*storageAlias
	}{
		Inner:        r.inner,
		Arrays:       r.arrays,
		storageAlias: (*storageAlias)(r),
	})
}

func (r *Storage) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Inner  map[string]*val  `json:"inner"`
		Arrays map[string][]int `json:"arrays"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	r.inner = aux.Inner
	r.arrays = aux.Arrays
	r.logger, _ = zap.NewProduction(zap.IncreaseLevel(zapcore.DPanicLevel))
	return nil
}

func (r *Storage) SaveToFile(filename string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("error marshalling storage: %v", err)
	}
	if err := ioutil.WriteFile(filename, data, 0666); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	r.logger.Info("Storage saved to file", zap.String("filename", filename))
	return nil
}

func (r *Storage) LoadFromFile(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	if err := json.Unmarshal(file, r); err != nil {
		return fmt.Errorf("error unmarshalling storage: %v", err)
	}
	r.logger.Info("Storage loaded from file", zap.String("filename", filename))
	return nil
}
