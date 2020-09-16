package main

import (
	"context"
	"errors"
)

var (
	ErrIterStopped = errors.New("error iter stopped")
)

// Callback
func IterCb(from, to int, cb func(int) error) error {
	for i := from; i < to; i++ {
		if err := cb(i); err != nil {
			return err
		}
	}
	return nil
}

// Closure
func IterClosure(from, to int) func() (int, error) {
	i := from
	return func() (int, error) {
		if i >= to {
			return 0, ErrIterStopped
		}
		curr := i
		i++
		return curr, nil
	}
}

// Chan
func IterChan(from, to int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := from; i < to; i++ {
			c <- i
		}
	}()
	return c
}

// Better Chan with context
func IterChanWithCtx(ctx context.Context, from, to int) <-chan int {
	c := make(chan int)
	go func() {
		defer close(c)
		for i := from; i < to; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				c <- i
			}
		}
	}()
	return c
}

// Better Chan with close
func IterChanWithClose(from, to int) (<-chan int, func()) {
	c := make(chan int)
	go func() {
		for i := from; i < to; i++ {
			c <- i
		}
	}()
	return c, func() { close(c) }
}

// State
type IterCur struct {
	current int
	to      int
}

func NewIterCur(from, to int) *IterCur {
	return &IterCur{
		current: from,
		to:      to,
	}
}

func (i *IterCur) Next() bool {
	if i.current >= i.to {
		return false
	}
	i.current++
	return true
}

func (i *IterCur) Current() int {
	return i.current
}

func main() {}
