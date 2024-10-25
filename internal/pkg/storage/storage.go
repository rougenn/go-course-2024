package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type kind string

type val struct {
	value_type      kind
	string_value    string
	int_value       int
	expiration_time int64
}

type arrVal struct {
	arr             []int
	expiration_time int64
}

const (
	KindString    = kind("S")
	KindInt       = kind("D")
	KindUndefined = kind("UNDEFINED")
)

func newArr() *arrVal {

}

type Storage struct {
	inner         map[string]*val    `json:"inner"`
	arrays        map[string]*arrVal `json:"arrays"`
	logger        *zap.Logger        `json:"-"`
	cleanDuration time.Duration      `json:"-`
	saveDuration  time.Duration      `json:"-`
	filename      string             `json:"-`
}

// Func creates a new storage with saving and cleaning duration time is seconds.
// It saves current version of storage to file.json every (cleanDuration) seconds
func NewStorage(saveDuration, cleanDuration time.Duration, filename string) (*Storage, error) {
	// to turn off the logger while benchmarks
	// logger, _ := zap.NewProduction(zap.IncreaseLevel(zapcore.DPanicLevel))

	logger, _ := zap.NewProduction()

	// defer logger.Sync()

	logger.Info("new storage created", zap.Int64("clean duration", int64(cleanDuration)),
		zap.Int64("save duration", int64(saveDuration)))

	r := Storage{
		inner:         make(map[string]*val),
		logger:        logger,
		arrays:        make(map[string]*arrVal),
		cleanDuration: cleanDuration,
		saveDuration:  saveDuration,
		filename:      filename,
	}

	r.RunGarbageCollector()
	r.RunStorageSaving()

	return &r, nil
}

func (r *Storage) RunGarbageCollector() {
	ticker := time.NewTicker(r.cleanDuration)

	go func() {
		for range ticker.C {
			// r.logger.Info("garbage collector is running")
			r.GarbageCollect()
		}
	}()
}

func (r *Storage) GarbageCollect() {
	curTime := time.Now().Unix()
	for key, v := range r.inner {
		if v.expiration_time != 0 && v.expiration_time < curTime {
			delete(r.inner, key)
			r.logger.Info("deleted expired key", zap.String("key", key))
		}
	}
}

func (r *Storage) RunStorageSaving() {
	ticker := time.NewTicker(r.saveDuration)

	go func() {
		for range ticker.C {
			r.SaveToFile(r.filename)
		}
	}()
}

// Set устанавливает значение по указанному ключу с опциональным временем истечения.
// Время истечения указывается в секундах. Например, для установки времени истечения на 5 минут,
// передайте 300 (5 минут * 60 секунд).
func (r *Storage) Set(key string, input_val interface{}, expiration_seconds ...int64) error {
	t := int64(0)
	switch len(expiration_seconds) {
	case 1:
		if expiration_seconds[0] > 0 {
			t = expiration_seconds[0] + time.Now().Unix()
		}
		if expiration_seconds[0] < 0 {
			return ErrIncorrectArgs
		}
	case 0:
		t = 0
	default:
		r.logger.Error("incorrect args")
		return ErrIncorrectArgs
	}

	// defer r.logger.Sync()
	if _, exists := r.arrays[key]; exists {
		r.logger.Error("по данному ключу существует значение другого типа")
		return ErrKeyAlreadyExists
	}

	switch v := input_val.(type) {
	case int:
		r.inner[key] = &val{
			value_type:      KindInt,
			int_value:       v,
			expiration_time: t,
		}
		r.logger.Info("key obtained", zap.String("key", key),
			zap.Int("val", v), zap.String("type", string(KindInt)))
	case string:
		r.inner[key] = &val{
			value_type:      KindString,
			string_value:    v,
			expiration_time: t,
		}
		r.logger.Info("key obtained", zap.String("key", key),
			zap.String("val", v), zap.String("type", string(KindString)))
	default:
		r.logger.Error("unsupported value type")
		return ErrUnsupportedValueType
	}
	return nil
}

var (
	ErrKeyDoesntExist       = errors.New("key value doesnt exist")
	ErrIndexOutOfRange      = errors.New("index is out of range")
	ErrKeyAlreadyExists     = errors.New("key already exists")
	ErrIncorrectArgs        = errors.New("function got incorrect arguments")
	ErrUnsupportedValueType = errors.New("unsupported value type")
)

func (r *Storage) GetValue(key string) (*val, error) {
	// defer r.logger.Sync()
	curTime := time.Now().Unix()
	val, ok := r.inner[key]

	if !ok || (val.expiration_time != 0 && val.expiration_time < curTime) {
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

	r.arrays[key].arr = append(r.arrays[key].arr, arr...)

	r.logger.Info("New elems added to RIGHT side of slice",
		zap.Int("count of elems", len(arr)), zap.String("key", key))
}

func (r *Storage) Lpush(key string, arr ...int) {

	r.arrays[key].arr = append(arr, r.arrays[key].arr...)

	r.logger.Info("New elems added to LEFT side of slice",
		zap.Int("count of elems", len(arr)), zap.String("key", key))
}

func (r *Storage) Raddtoset(key string, arr ...int) {

	for _, elem := range arr {
		exists := false
		for _, i := range r.arrays[key].arr {
			if elem == i {
				exists = true
				break
			}
		}
		if !exists {
			r.arrays[key].arr = append(r.arrays[key].arr, elem)
		}
	}
	r.logger.Info("New elements added", zap.String("key", key))
}

func (r *Storage) DeleteSegment(key string, l int, ri int) ([]int, error) {
	leng := len(r.arrays[key].arr)

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

	left_part := r.arrays[key].arr[:l]
	right_part := r.arrays[key].arr[ri+1:]

	deleted := make([]int, ri-l+1)
	copy(deleted, r.arrays[key].arr[l:ri+1])
	// fmt.Println(left_part, deleted, right_part)

	r.arrays[key].arr = append(left_part, right_part...)
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

	length := len(r.arrays[key].arr)
	switch le := len(args); le {
	case 0:
		cnt := 1

		if cnt > len(r.arrays[key].arr) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := r.arrays[key].arr[:cnt]

		r.arrays[key].arr = r.arrays[key].arr[cnt:]
		r.logger.Info("deleted elems from left",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 1:
		cnt := args[0]

		if cnt > len(r.arrays[key].arr) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := make([]int, cnt)
		copy(deleted, r.arrays[key].arr[:cnt])

		r.arrays[key].arr = r.arrays[key].arr[cnt:]
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
	length := len(r.arrays[key].arr)
	switch le := len(args); le {
	case 0:
		cnt := 1

		if cnt > len(r.arrays[key].arr) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := make([]int, cnt)
		// fmt.Println(length, cnt, r.arrays[key])
		copy(deleted, r.arrays[key].arr[length-cnt:length])

		r.arrays[key].arr = r.arrays[key].arr[:length-cnt]
		r.logger.Info("deleted elems from right",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 1:

		cnt := args[0]

		if cnt > len(r.arrays[key].arr) {
			return []int{length}, ErrIndexOutOfRange
		}
		deleted := make([]int, cnt)
		copy(deleted, r.arrays[key].arr[length-cnt:length])

		r.arrays[key].arr = r.arrays[key].arr[:length-cnt]
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
	if _, exists := r.inner[key]; exists {
		r.logger.Error("по данному ключу существует значение другого типа")
		return ErrKeyAlreadyExists
	}
	arr, ok := r.arrays[key]
	if !ok {
		r.logger.Error(ErrKeyDoesntExist.Error())
		return ErrKeyDoesntExist
	}
	if index >= len(arr.arr) {
		r.logger.Error(ErrIndexOutOfRange.Error())
		return ErrIndexOutOfRange
	}
	arr.arr[index] = new_val
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
	if index >= len(arr.arr) {
		r.logger.Error(ErrIndexOutOfRange.Error())
		return 0, ErrIndexOutOfRange
	}
	r.logger.Info("value requested", zap.String("key", key),
		zap.Int("index", index))
	return arr.arr[index], nil
}

func (av *arrVal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Arr             []int `json:"int array"`
		Expiration_time int64 `json:"expiration time"`
	}{
		Arr:             av.arr,
		Expiration_time: av.expiration_time,
	})
}

func (av *arrVal) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Arr             []int `json:"int array"`
		Expiration_time int64 `json:"expiration time"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	av.arr = aux.Arr
	av.expiration_time = aux.Expiration_time

	return nil
}

func (v *val) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ValueType       kind   `json:"value_type"`
		StringValue     string `json:"string_value"`
		IntValue        int    `json:"int_value"`
		Expiration_time int64  `json:"expiration time"`
	}{
		ValueType:       v.value_type,
		StringValue:     v.string_value,
		IntValue:        v.int_value,
		Expiration_time: v.expiration_time,
	})
}

func (v *val) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ValueType       kind   `json:"value_type"`
		StringValue     string `json:"string_value"`
		IntValue        int    `json:"int_value"`
		Expiration_time int64  `json:"expiration time"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	v.value_type = aux.ValueType
	v.string_value = aux.StringValue
	v.int_value = aux.IntValue
	v.expiration_time = aux.Expiration_time

	return nil
}

func (r *Storage) MarshalJSON() ([]byte, error) {
	type storageAlias Storage
	return json.Marshal(&struct {
		Inner  map[string]*val    `json:"inner"`
		Arrays map[string]*arrVal `json:"arrays"`
		*storageAlias
	}{
		Inner:        r.inner,
		Arrays:       r.arrays,
		storageAlias: (*storageAlias)(r),
	})
}

func (r *Storage) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Inner  map[string]*val    `json:"inner"`
		Arrays map[string]*arrVal `json:"arrays"`
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

	temp := "temp_" + filename
	if err := ioutil.WriteFile(temp, data, 0666); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	if err := os.Rename(temp, filename); err != nil {
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

func (r *Storage) Expire(key string, ex int64) {

}
