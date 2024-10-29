package storage

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	key   string
	value string
	kind  kind
}

type benchCase struct {
	cnt int
}

var cases = []benchCase{
	{1},
	{10},
	{100},
	{1000},
	{10000},
}

func BenchmarkSet(b *testing.B) {
	for in, tCase := range cases {
		b.Run(strconv.Itoa(in), func(b *testing.B) {
			s, _ := NewStorage(1000, 10000, "test.json")

			for i := 0; i < tCase.cnt; i++ {
				s.Set(strconv.Itoa(i), strconv.Itoa(i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Set(strconv.Itoa(rand.Intn(tCase.cnt)), "fkjashdf")
			}
		})
	}
}

//                                              results:
// goos: linux
// goarch: amd64
// pkg: hw1/internal/pkg/storage
// cpu: AMD Ryzen 7 5700U with Radeon Graphics
// BenchmarkSet/1-16  	 4511271	       254.2 ns/op	     296 B/op	       4 allocs/op
// BenchmarkSet/10-16 	 4381594	       264.9 ns/op	     296 B/op	       4 allocs/op
// BenchmarkSet/100-16         	 4434793	       268.9 ns/op	     296 B/op	       4 allocs/op
// BenchmarkSet/1000-16        	 3699061	       325.0 ns/op	     303 B/op	       4 allocs/op
// BenchmarkSet/10000-16       	 2532891	       479.5 ns/op	     303 B/op	       4 allocs/op
// PASS
// ok  	hw1/internal/pkg/storage	7.592s

func BenchmarkGet(b *testing.B) {
	for in, tCase := range cases {
		b.Run(strconv.Itoa(in), func(b *testing.B) {
			s, _ := NewStorage(1000, 10000, "test.json")

			for i := 0; i < tCase.cnt; i++ {
				s.Set(strconv.Itoa(i), strconv.Itoa(i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Get(strconv.Itoa(rand.Intn(tCase.cnt)))
			}
		})
	}
}

//                                              results:
// goos: linux
// goarch: amd64
// pkg: hw1/internal/pkg/storage
// cpu: AMD Ryzen 7 5700U with Radeon Graphics
// BenchmarkGet/1-16     	 7449374	       160.6 ns/op	     208 B/op	       2 allocs/op
// BenchmarkGet/10-16    	 6576946	       186.5 ns/op	     208 B/op	       2 allocs/op
// BenchmarkGet/100-16   	 6255698	       190.3 ns/op	     208 B/op	       2 allocs/op
// BenchmarkGet/1000-16  	 4870556	       250.5 ns/op	     213 B/op	       3 allocs/op
// BenchmarkGet/10000-16 	 3803904	       338.5 ns/op	     215 B/op	       3 allocs/op
// PASS
// ok  	hw1/internal/pkg/storage	16.542s

func BenchmarkGetSet(b *testing.B) {
	for in, tCase := range cases {
		b.Run(strconv.Itoa(in), func(b *testing.B) {
			s, _ := NewStorage(1000, 10000, "test.json")

			for i := 0; i < tCase.cnt; i++ {
				s.Set(strconv.Itoa(i), strconv.Itoa(i))
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s.Set(strconv.Itoa(i), strconv.Itoa(i))
				s.Get(strconv.Itoa(rand.Intn(tCase.cnt)))
			}
		})
	}
}

//                                              results:
// goos: linux
// goarch: amd64
// pkg: hw1/internal/pkg/storage
// cpu: AMD Ryzen 7 5700U with Radeon Graphics
// BenchmarkGetSet/1-16  	 1320218	       873.4 ns/op	     557 B/op	       6 allocs/op
// BenchmarkGetSet/10-16 	 1407057	       815.4 ns/op	     552 B/op	       6 allocs/op
// BenchmarkGetSet/100-16         	 1000000	      1003 ns/op	     587 B/op	       6 allocs/op
// BenchmarkGetSet/1000-16        	 1000000	      1010 ns/op	     591 B/op	       7 allocs/op
// BenchmarkGetSet/10000-16       	 1000000	      1105 ns/op	     592 B/op	       8 allocs/op
// PASS
// ok  	hw1/internal/pkg/storage	7.330s

func TestSetGet(t *testing.T) {
	cases := []testCase{
		{"hello", "world", KindString},
		{"jslad", "a", KindString},
		{"hello", "123", KindInt},
		{"aslkdjdflska", "fsdf", KindString},
		{"asldkfjasldfkjasl;dfjal;kdfjakl;dfjkjssadjfla", "wor2", KindString},
		{"gg", "gg", KindString},
		{"123", "321", KindInt},
		{"j", "1", KindInt},
		{"2", "1", KindInt},
	}

	s, _ := NewStorage(100000, 10000, "test.json")

	for in, c := range cases {
		t.Run(strconv.Itoa(in), func(t *testing.T) {
			s.Set(c.key, c.value)
			sValue, err := s.Get(c.key)
			if err != nil {
				t.Errorf("ошибка бенчмарка: %v", err)
			}
			sKind, err := s.GetKind(c.key)
			if err != nil {
				t.Errorf("ошибка бенчмарка: %v", err)
			}

			if sValue != c.value {
				t.Errorf("values not equal: %v, %v", sValue, c.value)
			}

			if sKind != c.kind {
				t.Errorf("types are not equal: %v, and expected: %v", sKind, c.kind)
			}
		})
	}
}

func TestGarbageCollect(t *testing.T) {
	r, err := NewStorage(time.Minute*20, time.Minute*60, "test.json")

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove("test.json")

	r.Set("key1", "value1", 1)
	r.Set("key2", "value2", 12345)
	r.Set("key3", "1234", 1)
	time.Sleep(2 * time.Second)
	r.GarbageCollect()

	_, err1 := r.Get("key1")
	_, err2 := r.Get("key2")
	_, err3 := r.Get("key3")

	assert.False(t, err1 == nil, "key1 должен быть удален")
	assert.True(t, err2 == nil, "key2 должен оставаться")
	assert.False(t, err3 == nil, "key3 должен быть удален")
}

func TestDatabaseSaveAndLoadWithGarbageCollector(t *testing.T) {
	r, err := NewStorage(time.Minute*20, time.Minute*60, "my-storage.json")
	defer os.Remove("my-storage.json")

	if err != nil {
		t.Fatal(err)
	}

	r.Set("key1", "value1", 2)
	r.Set("key2", "value2", 5)
	r.Set("key3", "value3", 10)

	r.SaveToFile("my-storage.json")

	r2, err := NewStorage(15*time.Minute, 1*time.Minute, "my-storage.json")
	if err != nil {
		t.Fatal(err)
	}
	r2.LoadFromFile("my-storage.json")

	time.Sleep(6 * time.Second)

	r2.GarbageCollect()

	_, err1 := r2.Get("key1")
	_, err2 := r2.Get("key2")
	_, err3 := r2.Get("key3")

	assert.False(t, err1 == nil, "key1 должен быть удален после garbage collection")
	assert.False(t, err2 == nil, "key2 должен быть удален после garbage collection")
	assert.True(t, err3 == nil, "key3 должен оставаться")
}
