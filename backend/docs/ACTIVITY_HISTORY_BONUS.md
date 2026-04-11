# Activity History System - Bonus: Async/Event-Based Extension

This document outlines how to extend the Activity History system to use an asynchronous, event-driven architecture for better performance and scalability.

## Current Implementation (Synchronous)

The current implementation logs activities synchronously within each service method:

```go
// Current approach - synchronous logging
func (s *TaskService) CreateTask(...) {
    // ... task creation logic ...
    
    // Log activity (blocks until complete)
    s.activitySvc.LogTaskCreated(projectID, userID, "", taskID, title)
}
```

## Problem with Synchronous Logging

1. **Latency Impact**: Each activity log adds latency to the main request
2. **Database Connection Pool**: Activity logging consumes connection pool resources
3. **Error Propagation**: Failures in logging can affect the main operation
4. **Scalability**: Under high load, logging becomes a bottleneck

## Solution: Event-Driven Architecture

### Architecture Overview

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   Service   │────▶│ Event Channel │────▶│ Activity Worker │
│  (Producer) │     │  (Buffered)  │     │   (Consumer)    │
└─────────────┘     └──────────────┘     └─────────────────┘
                                                   │
                                                   ▼
                                          ┌─────────────────┐
                                          │  Activity Repo   │
                                          │  (Database)      │
                                          └─────────────────┘
```

### Implementation Steps

#### 1. Create Event Types

```go
// internal/events/activity_events.go
package events

import "backend/internal/models"

type ActivityEvent struct {
    ProjectID   int64
    UserID     string
    UserName   string
    Action     models.ActivityAction
    EntityType models.EntityType
    EntityID   int64
    Details    string
}

type EventHandler interface {
    Handle(event ActivityEvent) error
}
```

#### 2. Create Event Channel Manager

```go
// internal/events/event_bus.go
package events

import (
    "log"
    "sync"
)

type EventBus struct {
    channel chan ActivityEvent
    handlers []EventHandler
    wg sync.WaitGroup
}

func NewEventBus(bufferSize int) *EventBus {
    return &EventBus{
        channel: make(chan ActivityEvent, bufferSize),
        handlers: make([]EventHandler, 0),
    }
}

func (eb *EventBus) RegisterHandler(h EventHandler) {
    eb.handlers = append(eb.handlers, h)
}

func (eb *EventBus) Publish(event ActivityEvent) {
    eb.channel <- event
}

func (eb *EventBus) Start(workers int) {
    for i := 0; i < workers; i++ {
        eb.wg.Add(1)
        go eb.worker()
    }
}

func (eb *EventBus) worker() {
    defer eb.wg.Done()
    for event := range eb.channel {
        for _, handler := range eb.handlers {
            if err := handler.Handle(event); err != nil {
                log.Printf("Error handling event: %v", err)
            }
        }
    }
}

func (eb *EventBus) Stop() {
    close(eb.channel)
    eb.wg.Wait()
}
```

#### 3. Create Activity Event Handler

```go
// internal/events/activity_handler.go
package events

import "backend/internal/services"

type ActivityEventHandler struct {
    activityService *services.ActivityService
}

func NewActivityEventHandler(svc *services.ActivityService) *ActivityEventHandler {
    return &ActivityEventHandler{activityService: svc}
}

func (h *ActivityEventHandler) Handle(event ActivityEvent) error {
    return h.activityService.LogActivity(
        event.ProjectID,
        event.UserID,
        event.UserName,
        event.Action,
        event.EntityType,
        event.EntityID,
        event.Description,
        event.Details,
    )
}
```

#### 4. Update Activity Service for Async Mode

```go
// internal/services/activity_service.go
type ActivityService struct {
    db            *sql.DB
    activityRepo  *repository.ActivityRepository
    memberService *ProjectMemberService
    eventBus      *events.EventBus
    asyncMode     bool
}

func NewActivityService(db *sql.DB, memberService *ProjectMemberService, eventBus *events.EventBus) *ActivityService {
    return &ActivityService{
        db:            db,
        activityRepo:  repository.NewActivityRepository(db),
        memberService: memberService,
        eventBus:      eventBus,
        asyncMode:     eventBus != nil,
    }
}

func (s *ActivityService) LogActivityAsync(event events.ActivityEvent) {
    if s.asyncMode && s.eventBus != nil {
        s.eventBus.Publish(event)
    }
}
```

#### 5. Update Task Service to Use Async Logging

```go
// internal/services/task_service.go
func (s *TaskService) CreateTask(...) {
    // ... existing logic ...
    
    // Async logging
    if s.activitySvc != nil {
        s.activitySvc.LogActivityAsync(events.ActivityEvent{
            ProjectID:   projectID,
            UserID:     userID,
            Action:     models.ActivityTaskCreated,
            EntityType: models.EntityTask,
            EntityID:   taskID,
            Details:    title,
        })
    }
}
```

### Benefits of Event-Driven Architecture

1. **Non-blocking**: Main request completes without waiting for logging
2. **Backpressure Handling**: Buffered channel prevents system overload
3. **Retry Logic**: Failed events can be retried automatically
4. **Event Sourcing Ready**: Events can be replayed for debugging
5. **Scalability**: Multiple workers can process events in parallel

### Configuration Options

```go
type ActivityConfig struct {
    AsyncEnabled bool
    BufferSize   int
    Workers      int
    RetryCount   int
}

// Usage
eventBus := events.NewEventBus(1000) // 1000 event buffer
eventBus.RegisterHandler(NewActivityEventHandler(activityService))
eventBus.Start(4) // 4 worker goroutines

activityService := NewActivityService(db, memberService, eventBus)
```

### Monitoring & Observability

Add metrics to track event processing:

```go
type EventMetrics struct {
    Published   prometheus.Counter
    Processed   prometheus.Counter
    Failed      prometheus.Counter
    Duration    prometheus.Histogram
}
```

### Database Connection Optimization

With async logging, database connection usage becomes more predictable:

- **Sync Mode**: Uses connections proportional to request load
- **Async Mode**: Uses fixed worker pool for logging

```go
// Separate connection pool for activity logging
dbActivity, _ := sql.Open("sqlite3", "taskify.db")
dbActivity.SetMaxOpenConns(5) // Dedicated pool for activity logs

activityRepo := repository.NewActivityRepository(dbActivity)
```

## Migration Path

1. **Phase 1**: Add async mode as opt-in feature
2. **Phase 2**: Add monitoring and tune buffer sizes
3. **Phase 3**: Make async the default mode
4. **Phase 4**: Add event sourcing capabilities

## Testing Async Implementation

```go
func TestAsyncActivityLogging(t *testing.T) {
    eventBus := events.NewEventBus(100)
    handler := &mockHandler{events: make([]events.ActivityEvent, 0)}
    eventBus.RegisterHandler(handler)
    eventBus.Start(1)
    defer eventBus.Stop()
    
    eventBus.Publish(events.ActivityEvent{
        ProjectID: 1,
        UserID:   "user123",
        Action:   models.ActivityTaskCreated,
    })
    
    // Wait for event to be processed
    time.Sleep(100 * time.Millisecond)
    
    assert.Len(t, handler.events, 1)
}
```

## Conclusion

The event-driven extension provides:
- Improved request latency
- Better resource utilization  
- Built-in scalability path
- Enhanced observability

The current synchronous implementation is production-ready for moderate loads. The async extension is recommended for high-traffic deployments.
