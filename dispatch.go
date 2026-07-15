package omnitui

// Event queueing is coordinated by App.cycle and the dispatcher in app.go.
// Keeping queue ownership in the runtime goroutine makes state updates from
// handlers and external goroutines deterministic.
