package main

import (
	"testing"
)

func BenchmarkIterCb(b *testing.B) {
	for n := 0; n < b.N; n++ {
		IterCb(0, 1000, func(x int) error { return nil })
	}
}

func BenchmarkIterClosure(b *testing.B) {
	for n := 0; n < b.N; n++ {
		next := IterClosure(0, 1000)
		for _, err := next(); err != nil; _, err = next() {
		}
	}
}

func BenchmarkIterChan(b *testing.B) {
	for n := 0; n < b.N; n++ {
		iter := IterChan(0, 1000)
		for range iter {
		}
	}
}

func BenchmarkIterCur(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cur := NewIterCur(0, 1000)
		for cur.Next() {
			cur.Current()
		}
	}
}
