package ee

import (
	"context"
	"sync"
	"testing"
	"time"
)

type EventA struct {
	A string
}

type EventB struct {
	B string
}

func TestEventEmitter(t *testing.T) {
	ctx := context.TODO()
	defer ctx.Done()

	ee := NewEventEmitter(ctx)

	if ee.size != 10 {
		t.Errorf("Expected 10 got %d", ee.size)
	}
}

func TestEventEmitterSize(t *testing.T) {
	ctx := context.TODO()
	defer ctx.Done()

	ee := NewEventEmitter(ctx, 5)

	if ee.size != 5 {
		t.Errorf("Expected 5 got %d", ee.size)
	}
}

func TestEventEmittedSubscribe(t *testing.T) {
	ctx := context.TODO()
	defer ctx.Done()

	ee := NewEventEmitter(ctx)

	var wg sync.WaitGroup

	wg.Add(1)

	ch := ee.Subscribe(EventA{})

	go func() {
		defer wg.Done()

		val := <-ch

		if val.(EventA).A != "a" {
			t.Errorf("Expected a got %s", val.(EventA).A)
		}
	}()

	ee.Emit(EventA{A: "a"})
	wg.Wait()
}

func TestEventEmittedSubscribeSafeExit(t *testing.T) {
	ctx, close := context.WithTimeout(context.TODO(), 2*time.Second)
	defer close()

	ee := NewEventEmitter(ctx)

	var wg sync.WaitGroup

	count := 0

	wg.Add(1)

	ch := ee.Subscribe(EventA{})

	go func() {
		defer wg.Done()

		for {
			val, ok := <-ch

			if !ok {
				return
			}

			count += 1

			if val.(EventA).A != "a" {
				t.Errorf("Expected a got %s", val.(EventA).A)
			}
		}
	}()

	ee.Emit(EventA{A: "a"})
	ee.Emit(EventA{A: "a"})
	ee.Emit(EventB{B: "b"})

	if len(ee.subscribers) != 1 {
		t.Errorf("Expected 1 got %d", len(ee.subscribers))
	}

	time.Sleep(1 * time.Second)

	close()

	time.Sleep(1 * time.Second)

	wg.Wait()

	if len(ee.subscribers) != 0 {
		t.Errorf("Expected 0 got %d", len(ee.subscribers))
	}

	if count != 2 {
		t.Errorf("Expected function to be called 2 times got %d", count)
	}

}

func TestEventEmittedUnsubscribe(t *testing.T) {
	ctx, close := context.WithTimeout(context.TODO(), 2*time.Second)
	defer close()

	ee := NewEventEmitter(ctx)

	var wg sync.WaitGroup

	count := 0

	wg.Add(1)

	c, cancel := context.WithCancel(context.TODO())

	ch := ee.Subscribe(EventA{}, c)

	go func() {
		defer wg.Done()

		for {
			val, ok := <-ch

			if !ok {
				return
			}

			count += 1

			if val.(EventA).A != "a" {
				t.Errorf("Expected a got %s", val.(EventA).A)
			}
		}
	}()

	ee.Emit(EventA{A: "a"})
	ee.Emit(EventB{B: "b"})

	time.Sleep(1 * time.Second)

	cancel()

	time.Sleep(1 * time.Second)

	ee.Emit(EventA{A: "a"})

	wg.Wait()

	if len(ee.subscribers) != 0 {
		t.Errorf("Expected 0 got %d", len(ee.subscribers))
	}

	if count != 1 {
		t.Errorf("Expected function to be called onces got %d", count)
	}

}

func TestEventEmittedOnce(t *testing.T) {
	ctx, close := context.WithTimeout(context.TODO(), 5*time.Second)
	defer close()

	ee := NewEventEmitter(ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	var val EventA

	go func() {
		defer wg.Done()
		ok := ee.Once(&val)

		if !ok {
			t.Errorf("Expected ee.once to read a value")
		}

		if val.A != "a" {
			t.Errorf("Expected a got %s", val.A)
		}
	}()

	time.Sleep(1 * time.Second)

	ee.Emit(EventA{A: "a"})

	wg.Wait()

	time.Sleep(1 * time.Second)

	if val.A != "a" {
		t.Errorf("Expected a got %s", val.A)
	}

	if len(ee.subscribers) != 0 {
		t.Errorf("Expected 0 got %d", len(ee.subscribers))
	}
}

func TestEventEmittedOnceEarlyExit(t *testing.T) {
	ctx, close := context.WithTimeout(context.TODO(), 5*time.Second)
	defer close()

	ee := NewEventEmitter(ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	var val EventA

	earlyCtx, earlyClose := context.WithCancel(context.TODO())

	go func() {
		defer wg.Done()
		ok := ee.Once(&val, earlyCtx)

		if ok {
			t.Errorf("Expected ee.once to not be called")
		}
	}()

	time.Sleep(1 * time.Second)

	earlyClose()

	time.Sleep(1 * time.Second)

	ee.Emit(EventA{A: "a"})

	time.Sleep(1 * time.Second)

	wg.Wait()

	time.Sleep(1 * time.Second)

	if len(ee.subscribers) != 0 {
		t.Errorf("Expected 0 got %d", len(ee.subscribers))
	}
}
