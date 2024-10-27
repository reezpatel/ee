package ee

import (
	"context"
	"reflect"
	"sync"
)

type EventSubscriber struct {
	ch      chan interface{}
	shape   reflect.Type
	context *context.Context
}

type EventEmitter struct {
	subscribers map[*EventSubscriber]interface{}
	mu          sync.Mutex
	context     context.Context
	size        int
}

func NewEventEmitter(context context.Context, bufferSize ...int) *EventEmitter {
	s := 10

	if len(bufferSize) > 0 {
		s = bufferSize[0]
	}

	e := &EventEmitter{
		context:     context,
		size:        s,
		subscribers: make(map[*EventSubscriber]interface{}),
	}

	return e
}

func (e *EventEmitter) Subscribe(t interface{}, context ...context.Context) <-chan interface{} {
	ctx := &e.context

	if len(context) > 0 {
		ctx = &context[0]
	}

	s := &EventSubscriber{
		ch:      make(chan interface{}, e.size),
		shape:   reflect.TypeOf(t),
		context: ctx,
	}

	e.addSubscriber(s)

	return s.ch
}

func (e *EventEmitter) addSubscriber(s *EventSubscriber) {
	e.mu.Lock()
	defer e.mu.Unlock()

	go func(e *EventEmitter, s *EventSubscriber) {

		defer func() {
			e.mu.Lock()

			defer e.mu.Unlock()
			e.unsubscribe(s)
		}()

		select {
		case <-(*s.context).Done():
		case <-e.context.Done():
		}
	}(e, s)

	e.subscribers[s] = struct{}{}

}

func (e *EventEmitter) unsubscribe(s *EventSubscriber) {
	close(s.ch)
	delete(e.subscribers, s)
}

func (e *EventEmitter) Emit(msg interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	v := reflect.TypeOf(msg)

	for s := range e.subscribers {
		if ok := v.AssignableTo(s.shape); ok {
			s.ch <- msg
		}
	}
}

func (e *EventEmitter) Once(ptr interface{}, ctx ...context.Context) bool {
	c, cancel := context.WithCancel(context.TODO())
	defer cancel()

	exit := context.TODO()

	if len(ctx) > 0 {
		exit = ctx[0]
	}

	t := reflect.ValueOf(ptr).Elem()

	s := &EventSubscriber{
		ch:      make(chan interface{}, e.size),
		shape:   t.Type(),
		context: &c,
	}

	e.addSubscriber(s)

	select {
	case <-exit.Done():
		return false
	case val, ok := <-s.ch:
		if ok {
			t.Set(reflect.ValueOf(val))
		}
		return ok
	}
}
