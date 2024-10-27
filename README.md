# EventEmitter

A Simple EventEmitter for go.

> [!NOTE]
> I wrote this library to as part of learning Go, while is working on a pet project while involves heavy of socket and in-mem pusub. If there are features missing, please open an issue.
> Check the blog [here](https://reezpatel.medium.com/simple-event-emitter-in-go-c6b8f8d4c2f7) for implementation details.

## Highlights

- Simple & Clean API
- Thread Safe
- Buffering of events
- Dispatch multiple interfaces with single instance
- Listen to event once
- Ability to Unsubscribe from events

## Installation

```bash
go get github.com/reezpatel/ee
```

## Usage

For more example, checkout the `main_test.go` file.

```go
ctx, close := context.WithCancel(context.TODO())
defer close()

ee := NewEventEmitter(ctx)
ch := ee.Subscribe(EventA{}) // Subscribe to EventA
// Pass a cancellable context to Subscrible to close it early

// Event Listner
go func() {
    for {
        val, ok := <-ch

        if !ok {
            return
        }

        fmt.Println(val)
    }
}()


var val EventB

go func() {
    ok := ee.Once(&val) // Listen to EventB once
}()

ee.Emit(EventA{A: "a"}) // Dispatach different events
ee.Emit(EventB{B: "b"})
```

## API

### Create Instance

```go
// with default buffer size of 10 for each listner
ee := NewEventEmitter(ctx)

// with custom buffer size
ee := NewEventEmitter(ctx, 5)
```

### Dispatch Event

```go
ee.Emit(EventA{A: "a"})
```

### Subscribe to Event

#### Subscribe till the EventEmitter is closed

```go
ch := ee.Subscribe(EventA{})

go func() {
    for {
        val, ok := <-ch

        if !ok {
            // no more events
            return
        }
    }
}()
```

#### Subscribe with a cancellable context

```go
c, cancel := context.WithCancel(context.TODO())
ch := ee.Subscribe(EventB{}, ctx)

go func() {
    for {
        val, ok := <-ch

        if !ok {
            // no more events
            return
        }
    }
}()

...

cancel() // somewhere in the code
```

### Listen to Event once

#### Listen to Event till the EventEmitter is closed

```go
var val EventA
ok := ee.Once(&val)

if !ok {
    fmt.Println("No event was dispatched for EventA")
}
```

#### Listen to Event with a timeout

```go
var val EventA
ctx, close := context.WithTimeout(context.TODO(), 2*time.Second)
defer close()

ok := ee.Once(&val, ctx) // wait for EventA for 2 seconds

if !ok {
    fmt.Println("No event was dispatched for EventA")
}
```
