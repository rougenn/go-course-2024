package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

type kind string

type val struct {
	valueType   kind
	stringValue string
	intValue    int
}

const (
	KindString    = kind("S")
	KindInt       = kind("D")
	KindUndefined = kind("UNDEFINED")
)

var (
	ErrKeyDoesntExist       = errors.New("key value doesnt exist")
	ErrIndexOutOfRange      = errors.New("index is out of range")
	ErrKeyAlreadyExists     = errors.New("key already exists")
	ErrIncorrectArgs        = errors.New("function got incorrect arguments")
	ErrUnsupportedValueType = errors.New("unsupported value type")
)

type Storage struct {
	inner                 map[string]*val  `json:"inner"`
	arrays                map[string][]int `json:"arrays"`
	expirationTime        map[string]int64 `json:"expiration_time_in_milliseconds"`
	logger                *zap.Logger      `json:"-"`
	cleanDuration         time.Duration    `json:"-"`
	saveDuration          time.Duration    `json:"-"`
	filename              string           `json:"-"`
	closeStorageSaving    chan struct{}    `json:"-"`
	closeGarbageCollector chan struct{}    `json:"-"`
	wg                    *sync.WaitGroup  `json:"-"`
	mu                    *sync.Mutex      `json:"-"`
	db                    *sql.DB
}

const (
	CreateTable = `CREATE TABLE IF NOT EXISTS core (
		version bigserial PRIMARY KEY,
		timestamp bigint NOT NULL,
		payload JSONB NOT NULL
	)`
)

// Func creates a new storage with saving and cleaning duration time is seconds.
// It saves current version of storage to filename.json every {cleanDuration} seconds
func NewStorage(saveDuration, cleanDuration time.Duration, filename string) (*Storage, error) {
	// to turn off the logger while benchmarks
	// logger, _ := zap.NewProduction(zap.IncreaseLevel(zapcore.DPanicLevel))

	logger, _ := zap.NewProduction()

	logger.Info("new storage created", zap.Int64("clean duration", int64(cleanDuration)),
		zap.Int64("save duration", int64(saveDuration)))

	fmt.Println("Attempting to connect to PostgreSQL")
	db, err := sql.Open("postgres", "postgres://username:password@postgres:5432/storagedb?sslmode=disable")

	if err != nil {
		log.Fatal("connection:", err)
	}
	fmt.Println("Connected to PostgreSQL, pinging...")
	if err := db.Ping(); err != nil {
		log.Fatal("ping: ", err)
	}
	fmt.Println("Connected to PostgreSQL")

	// Создание таблицы при необходимости
	if _, err := db.Exec(CreateTable); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	var wg sync.WaitGroup
	r := Storage{
		inner:                 make(map[string]*val),
		logger:                logger,
		arrays:                make(map[string][]int),
		cleanDuration:         cleanDuration,
		saveDuration:          saveDuration,
		filename:              filename,
		expirationTime:        make(map[string]int64),
		closeGarbageCollector: make(chan struct{}),
		closeStorageSaving:    make(chan struct{}),
		wg:                    &wg,
		mu:                    &sync.Mutex{},
		db:                    db,
	}

	go r.RunStorageSaving(r.closeStorageSaving)
	go r.RunGarbageCollector(r.closeGarbageCollector)

	return &r, nil
}

func (r *Storage) RunGarbageCollector(closeChan chan struct{}) {
	for {
		select {
		case <-closeChan:
			return
		case <-time.After(r.cleanDuration):
			r.wg.Add(1)
			r.GarbageCollect()
		}
	}
}

func (r *Storage) Wait() {
	r.wg.Wait()
}

func (r *Storage) Stop() {
	r.closeGarbageCollector <- struct{}{}
	r.closeStorageSaving <- struct{}{}

	r.GarbageCollect()
	r.SaveToFile(r.filename)

	r.Wait()
}

func (r *Storage) GarbageCollect() {
	defer r.wg.Done()

	r.logger.Info("garbage collection started")

	curTime := time.Now().UnixMilli()
	expirationKeys := r.getRandomKeysWithExpiration(min(10, len(r.expirationTime)/5))

	for _, key := range expirationKeys {
		if expTime, exists := r.expirationTime[key]; exists && expTime != 0 && expTime < curTime {
			r.deleteKey(key)
		}
	}
}

func (r *Storage) getRandomKeysWithExpiration(count int) []string {
	keys := make([]string, 0, len(r.expirationTime))
	for key := range r.expirationTime {
		keys = append(keys, key)
	}

	if len(keys) <= count {
		return keys
	}

	randomKeys := make([]string, count)
	for i := 0; i < count; i++ {
		randomIndex := rand.Intn(len(keys))
		randomKeys[i] = keys[randomIndex]
		keys = append(keys[:randomIndex], keys[randomIndex+1:]...)
	}

	return randomKeys
}

func (r *Storage) deleteKey(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.inner[key]; exists {
		delete(r.inner, key)
		r.logger.Info("Deleted expired key from inner", zap.String("key", key))
	}
	if _, exists := r.arrays[key]; exists {
		delete(r.arrays, key)
		r.logger.Info("Deleted expired key from arrays", zap.String("key", key))
	}
	delete(r.expirationTime, key)
	r.logger.Info("Deleted expiration entry for key", zap.String("key", key))
}

func (r *Storage) CheckArrKey(key string) error {
	curTime := time.Now().UnixMilli()
	_, exists := r.arrays[key]

	if !exists {
		return ErrKeyDoesntExist
	}

	if r.expirationTime[key] != 0 && r.expirationTime[key] < curTime {
		delete(r.arrays, key)
		delete(r.expirationTime, key)

		return ErrKeyDoesntExist
	}
	return nil
}

func (r *Storage) RunStorageSaving(closeChan chan struct{}) {
	for {
		select {
		case <-closeChan:
			return
		case <-time.After(r.saveDuration):
			r.wg.Add(1)
			// r.SaveToFile(r.filename)
			if err := r.saveToPostgres(r.db); err != nil {
				r.logger.Error("error storage saving: " + err.Error())
			} else {
				r.logger.Info("successful saving storage to posrgres")
			}
		}
	}
}

func (r *Storage) Hset(args ...string) error {
	if len(args)%2 != 0 {
		return ErrIncorrectArgs
	}
	for i := 0; i < len(args); i += 2 {
		if err := r.Set(args[i], args[i+1]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Storage) Set(key string, inputVal string, expirationSeconds ...int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := int64(0)
	switch len(expirationSeconds) {
	case 1:
		if expirationSeconds[0] > 0 {
			t = time.Now().Add(time.Duration(expirationSeconds[0]) * time.Second).UnixMilli()
		}
		if expirationSeconds[0] < 0 {
			return ErrIncorrectArgs
		}
	case 0:
		t = 0
	default:
		r.logger.Error("incorrect args")
		return ErrIncorrectArgs
	}

	if _, exists := r.arrays[key]; exists {
		r.logger.Error("по данному ключу существует значение другого типа", zap.String("key", key))
		return ErrKeyAlreadyExists
	}

	intVal, err := strconv.Atoi(inputVal)
	if err == nil {
		r.inner[key] = &val{
			valueType: KindInt,
			intValue:  intVal,
		}
		r.expirationTime[key] = t

		r.logger.Info("key obtained", zap.String("key", key),
			zap.Int("val", intVal), zap.String("type", string(KindInt)))
		return nil
	}
	r.inner[key] = &val{
		valueType:   KindString,
		stringValue: inputVal,
	}
	r.expirationTime[key] = t

	r.logger.Info("key obtained", zap.String("key", key),
		zap.String("val", inputVal),
		zap.String("type", string(KindString)))
	return nil
}

func (r *Storage) GetValue(key string) (*val, error) {
	curTime := time.Now().UnixMilli()
	val, ok := r.inner[key]

	if !ok {
		r.logger.Info("key value doesnt exist", zap.String("key", key))
		return nil, ErrKeyDoesntExist
	}

	if r.expirationTime[key] != 0 && r.expirationTime[key] < curTime {
		delete(r.inner, key)
		delete(r.expirationTime, key)
		r.logger.Info("key value doesnt exist", zap.String("key", key))
		return nil, ErrKeyDoesntExist
	}

	if val.valueType == KindString {
		r.logger.Info("storage request", zap.String("key", key),
			zap.String("val", val.stringValue), zap.String("type", string(val.valueType)))
	} else {
		r.logger.Info("storage request", zap.String("key", key),
			zap.Int("val", val.intValue), zap.String("type", string(val.valueType)))
	}

	return val, nil
}

func (r *Storage) Get(key string) (string, error) {

	val, ok := r.GetValue(key)
	if ok != nil {
		return "", ok
	}
	switch val.valueType {
	case KindString:
		return val.stringValue, nil
	case KindInt:
		output := strconv.Itoa(val.intValue)
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
	return val.valueType, err
}

func (r *Storage) Rpush(key string, arr ...int) {
	if err := r.CheckArrKey(key); err != nil {
		r.expirationTime[key] = 0
	}

	r.arrays[key] = append(r.arrays[key], arr...)

	r.logger.Info("New elems added to RIGHT side of slice",
		zap.Int("count of elems", len(arr)), zap.String("key", key))
}

func (r *Storage) Lpush(key string, inputArr ...int) {
	if err := r.CheckArrKey(key); err != nil {
		r.expirationTime[key] = 0
	}

	r.arrays[key] = append(inputArr, r.arrays[key]...)

	r.logger.Info("New elems added to LEFT side of slice",
		zap.Int("count of elems", len(inputArr)), zap.String("key", key))
}

func (r *Storage) Raddtoset(key string, arr ...int) error {

	if err := r.CheckArrKey(key); err != nil {
		return err
	}

	for _, elem := range arr {
		exists := false
		for _, i := range r.arrays[key] {
			if elem == i {
				exists = true
				break
			}
		}
		if !exists {
			_, ex := r.arrays[key]
			if !ex {
				r.expirationTime[key] = 0
			}
			r.arrays[key] = append(r.arrays[key], elem)
		}
	}
	r.logger.Info("New elements added", zap.String("key", key))
	return nil
}

func (r *Storage) DeleteSegment(key string, l int, ri int) ([]int, error) {

	if err := r.CheckArrKey(key); err != nil {
		return nil, err
	}

	leng := len(r.arrays[key])

	if leng == 0 {
		return nil, nil
	}

	if ri < 0 {
		ri = leng + ri%leng
	}

	// fmt.Println(l, ri, r.arrays[key])

	if l > ri || l >= leng || ri >= leng {
		r.logger.Error("invalid indexes")
		return nil, ErrIndexOutOfRange
	}

	leftPart := r.arrays[key][:l]
	rightPart := r.arrays[key][ri+1:]

	deleted := make([]int, ri-l+1)
	copy(deleted, r.arrays[key][l:ri+1])

	r.arrays[key] = append(leftPart, rightPart...)
	r.logger.Info("Some elems has deleted from array",
		zap.String("key", key), zap.Int("left index", l),
		zap.Int("right index", ri))
	return deleted, nil
}

func (r *Storage) Lpop(key string, args ...int) ([]int, error) {
	if err := r.CheckArrKey(key); err != nil {
		return nil, err
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
		return nil, ErrIndexOutOfRange
	}
}

func (r *Storage) Rpop(key string, args ...int) ([]int, error) {
	if err := r.CheckArrKey(key); err != nil {
		return nil, err
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
		copy(deleted, r.arrays[key][length-cnt:length])

		r.arrays[key] = r.arrays[key][:length-cnt]
		r.logger.Info("deleted elems from right",
			zap.String("key", key), zap.Int("count", cnt))
		return deleted, nil
	case 2:
		return r.DeleteSegment(key, args[0], args[1])

	default:
		r.logger.Error("Invalid count of arguments, max count is 3")
		return nil, ErrIndexOutOfRange
	}
}

func (r *Storage) Lset(key string, index int, newVal int) error {
	if _, exists := r.inner[key]; exists {
		r.logger.Error("по данному ключу существует значение другого типа")
		return ErrKeyAlreadyExists
	}
	arr, ok := r.arrays[key]
	if !ok {
		r.logger.Error(ErrKeyDoesntExist.Error())
		return ErrKeyDoesntExist
	}
	if index >= len(arr) {
		r.logger.Error(ErrIndexOutOfRange.Error())
		return ErrIndexOutOfRange
	}
	arr[index] = newVal
	r.logger.Info("element changed", zap.String("key", key),
		zap.Int("index", index), zap.Int("new val", newVal))
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
		StringValue string `json:"string_value"`
		IntValue    int    `json:"int_value"`
	}{
		ValueType:   v.valueType,
		StringValue: v.stringValue,
		IntValue:    v.intValue,
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

	v.valueType = aux.ValueType
	v.stringValue = aux.StringValue
	v.intValue = aux.IntValue

	return nil
}

func (r *Storage) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Inner          map[string]*val  `json:"inner"`
		Arrays         map[string][]int `json:"arrays"`
		ExpirationTime map[string]int64
	}{
		Inner:          r.inner,
		Arrays:         r.arrays,
		ExpirationTime: r.expirationTime,
	})
}

func (r *Storage) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Inner          map[string]*val  `json:"inner"`
		Arrays         map[string][]int `json:"arrays"`
		ExpirationTime map[string]int64
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	r.inner = aux.Inner
	r.arrays = aux.Arrays
	r.expirationTime = aux.ExpirationTime
	r.logger, _ = zap.NewProduction()
	return nil
}

func (r *Storage) SaveToFile(filename string) error {

	defer r.wg.Done()

	data, err := json.Marshal(r)
	if err != nil {
		r.logger.Error("error marshalling storage: ", zap.String("filename", r.filename))
		return fmt.Errorf("error marshalling storage: %v", err)
	}

	temp := filepath.Join(filepath.Dir(filename), "storage_temp.json")

	if err := ioutil.WriteFile(temp, data, 0666); err != nil {
		fmt.Println("error writing to file: ", err)
		return fmt.Errorf("error writing to file: %v", err)
	}

	if err := os.Rename(temp, r.filename); err != nil {
		fmt.Println("error writing to file: ", err)
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

// ex = expiration time in seconds
func (r *Storage) Expire(key string, ex int64) bool {
	currentTime := time.Now().UnixMilli()
	if _, exists := r.inner[key]; exists &&
		(r.expirationTime[key] == 0 ||
			r.expirationTime[key] > currentTime) {

		if ex != 0 {
			r.expirationTime[key] = time.Now().Add(time.Duration(ex) * time.Second).UnixMilli()
		} else {
			r.expirationTime[key] = 0
		}
		return true
	}

	if _, exists := r.arrays[key]; exists &&
		(r.expirationTime[key] == 0 ||
			r.expirationTime[key] > currentTime) {

		if ex != 0 {
			r.expirationTime[key] = time.Now().Add(time.Duration(ex) * time.Second).UnixMilli()
		} else {
			r.expirationTime[key] = 0
		}
	}

	return false
}

func (r *Storage) saveToPostgres(db *sql.DB) error {
	// Преобразуем `state` в JSON
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}

	timestamp := time.Now().Unix()

	_, err = db.Exec(`INSERT INTO core (timestamp, payload) VALUES ($1, $2)`, timestamp, payload)
	if err != nil {
		return err
	}

	_, err = db.Exec(`DELETE FROM core WHERE version NOT IN (SELECT version FROM core ORDER BY version DESC LIMIT 5)`)
	if err != nil {
		return err
	}

	return nil
}

func (r *Storage) LoadFromPostgres() error {
	row := r.db.QueryRow(`SELECT payload FROM core ORDER BY version DESC LIMIT 1`)

	var payload []byte
	err := row.Scan(&payload)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info("Записей в базе данных не найдено, загружается пустое состояние.")
			return nil
		}
		return fmt.Errorf("ошибка при чтении последней версии состояния из базы данных: %v", err)
	}

	err = json.Unmarshal(payload, r)
	if err != nil {
		return fmt.Errorf("ошибка при декодировании JSON состояния: %v", err)
	}

	r.logger.Info("Состояние загружено из базы данных PostgreSQL")
	return nil
}
