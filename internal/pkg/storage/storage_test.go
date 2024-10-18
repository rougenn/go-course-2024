package storage

import (
	"math/rand"
	"strconv"
	"testing"
)

type testCase struct {
	name  string
	key   string
	value string
	kind  kind
}

type benchCase struct {
	name string
	cnt  int
	// key string
	// value string
	// kind string
}

var cases = []benchCase{
	{"1", 1},
	{"10", 10},
	{"100", 100},
	{"1000", 1000},
	{"10000", 10000},
}

func BenchmarkSet(b *testing.B) {
	for _, tCase := range cases {
		b.Run(tCase.name, func(b *testing.B) {
			s := NewStorage()

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

func BenchmarkGet(b *testing.B) {
	for _, tCase := range cases {
		b.Run(tCase.name, func(b *testing.B) {
			s := NewStorage()

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

func BenchmarkGetSet(b *testing.B) {
	for _, tCase := range cases {
		b.Run(tCase.name, func(b *testing.B) {
			s := NewStorage()

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

func TestSetGet(t *testing.T) {
	cases := []testCase{
		{"1", "hello", "world", KindString},
		{"2", "jslad", "a", KindString},
		{"3", "hello", "123", KindInt},
		{"4", "aslkdjdflska", "fsdf", KindString},
		{"5", "asldkfjasldfkjasl;dfjal;kdfjakl;dfjkjssadjfla", "wor2", KindString},
		{"6", "gg", "gg", KindString},
		{"7", "123", "321", KindInt},
		{"8", "j", "1", KindInt},
		{"9", "2", "1", KindInt},
	}

	s := NewStorage()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s.Set(c.key, c.value)
			sValue := s.Get(c.key)
			sKind := s.GetKind(c.key)

			if *sValue != c.value {
				t.Errorf("values not equal: %v, %v", *sValue, c.value)
			}

			if *sKind != c.kind {
				t.Errorf("types are not equal: %v, %v", *sKind, c.kind)
			}
		})
	}
}
