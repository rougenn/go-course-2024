package storage

import (
	"math/rand"
	"strconv"
	"testing"
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
