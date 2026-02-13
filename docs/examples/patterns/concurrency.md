# Go Concurrency Patterns for Real-Time Updates

This document provides comprehensive documentation on Go concurrency patterns specifically tailored for implementing a high-performance statusline with <10ms render latency.

## Table of Contents

1. [Core Concurrency Principles](#core-concurrency-principles)
2. [Goroutines and Channel Patterns](#goroutines-and-channel-patterns)
3. [Worker Pools for Parallel Widget Rendering](#worker-pools-for-parallel-widget-rendering)
4. [Context Propagation and Cancellation](#context-propagation-and-cancellation)
5. [Rate Limiting Patterns](#rate-limiting-patterns)
6. [Synchronization with sync Package](#synchronization-with-sync-package)
7. [Preventing Goroutine Leaks](#preventing-goroutine-leaks)
8. [Backpressure and Flow Control](#backpressure-and-flow-control)
9. [Real-Time Event Streaming Patterns](#real-time-event-streaming-patterns)
10. [Performance Optimization](#performance-optimization)

## Core Concurrency Principles

### 1. Don't communicate by sharing memory; share memory by communicating
- Use channels to pass data between goroutines
- Avoid shared state whenever possible
- When shared state is necessary, protect it with sync primitives

### 2. Goroutines are cheap, use them
- Goroutines have minimal overhead (~2KB stack)
- The Go scheduler efficiently manages thousands of goroutines
- Don't be afraid to spawn goroutines for concurrent work

### 3. Pipelines and Composition
- Build complex systems from simple, composable stages
- Each stage is a group of goroutines running the same function
- Channels connect stages, enabling clean data flow

## Goroutines and Channel Patterns

### Basic Channel Types

```go
// Unbuffered channel - synchronous communication
ch := make(chan int)

// Buffered channel - asynchronous up to buffer size
ch := make(chan int, 100)

// Directional channels for type safety
sendOnly := make(chan<- int)    // send-only
receiveOnly := make(<-chan int) // receive-only
```

### Pipeline Pattern

Essential for constructing streaming data pipelines where each stage processes data concurrently:

```go
// pipeline.go
package statusline

import (
    "context"
    "time"
)

// Stage 1: Generate status updates
func generateUpdates(ctx context.Context, interval time.Duration) <-chan StatusUpdate {
    out := make(chan StatusUpdate)
    
    go func() {
        defer close(out)
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                update := collectStatus()
                select {
                case out <- update:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    
    return out
}

// Stage 2: Enrich with additional data
func enrichStatus(ctx context.Context, in <-chan StatusUpdate) <-chan EnrichedStatus {
    out := make(chan EnrichedStatus)
    
    go func() {
        defer close(out)
        for update := range in {
            enriched := EnrichedStatus{
                Base:      update,
                Timestamp: time.Now(),
                Metadata:  collectMetadata(),
            }
            
            select {
            case out <- enriched:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}

// Stage 3: Format for display
func formatStatus(ctx context.Context, in <-chan EnrichedStatus) <-chan string {
    out := make(chan string)
    
    go func() {
        defer close(out)
        for status := range in {
            formatted := theme.Format(status)
            
            select {
            case out <- formatted:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}

// Usage: Chain stages together
func (s *StatusLine) startPipeline() {
    ctx := s.ctx
    updates := generateUpdates(ctx, s.interval)
    enriched := enrichStatus(ctx, updates)
    formatted := formatStatus(ctx, enriched)
    
    // Consume final output
    for output := range formatted {
        s.display(output)
    }
}
```

### Fan-Out/Fan-In Pattern

Distribute work across multiple workers and collect results:

```go
// fanout_fanin.go
package statusline

import (
    "context"
    "sync"
)

// FanOut distributes widget rendering across multiple workers
func fanOutWidgets(ctx context.Context, widgets []Widget, workers int) []<-chan RenderedWidget {
    // Create input channel
    in := make(chan Widget, len(widgets))
    
    // Send all widgets
    go func() {
        defer close(in)
        for _, w := range widgets {
            select {
            case in <- w:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    // Create worker channels
    outs := make([]<-chan RenderedWidget, workers)
    
    for i := 0; i < workers; i++ {
        out := make(chan RenderedWidget)
        outs[i] = out
        
        go func(out chan<- RenderedWidget) {
            defer close(out)
            for widget := range in {
                start := time.Now()
                content, err := widget.Render(ctx)
                
                result := RenderedWidget{
                    ID:       widget.ID(),
                    Content:  content,
                    Error:    err,
                    Duration: time.Since(start),
                }
                
                select {
                case out <- result:
                case <-ctx.Done():
                    return
                }
            }
        }(out)
    }
    
    return outs
}

// FanIn merges multiple channels into one
func fanInWidgets(ctx context.Context, channels ...<-chan RenderedWidget) <-chan RenderedWidget {
    out := make(chan RenderedWidget)
    var wg sync.WaitGroup
    
    // Start a goroutine for each input channel
    wg.Add(len(channels))
    for _, ch := range channels {
        go func(c <-chan RenderedWidget) {
            defer wg.Done()
            for result := range c {
                select {
                case out <- result:
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }
    
    // Close output when all inputs are done
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}

// Usage: Render widgets in parallel
func (s *StatusLine) renderWidgetsParallel(ctx context.Context) ([]RenderedWidget, error) {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
    defer cancel()
    
    // Fan out work
    workers := runtime.NumCPU()
    channels := fanOutWidgets(ctx, s.widgets, workers)
    
    // Fan in results
    results := fanInWidgets(ctx, channels...)
    
    // Collect all results
    var rendered []RenderedWidget
    for result := range results {
        if result.Error != nil {
            return nil, fmt.Errorf("widget %s: %w", result.ID, result.Error)
        }
        rendered = append(rendered, result)
    }
    
    return rendered, nil
}
```

### Tee-Channel Pattern

Split a single channel into multiple channels for parallel processing:

```go
// tee.go
package statusline

func teeStatus(ctx context.Context, in <-chan Status, n int) []<-chan Status {
    outs := make([]<-chan Status, n)
    
    for i := 0; i < n; i++ {
        out := make(chan Status)
        outs[i] = out
    }
    
    go func() {
        defer func() {
            for _, out := range outs {
                close(out)
            }
        }()
        
        for status := range in {
            // Send to all output channels
            for _, out := range outs {
                out := out // Capture for goroutine
                go func(s Status) {
                    select {
                    case out <- s:
                    case <-ctx.Done():
                    }
                }(status)
            }
        }
    }()
    
    return outs
}

// Usage: Process status in multiple ways simultaneously
func (s *StatusLine) processStatus(status <-chan Status) {
    // Split status stream
    streams := teeStatus(s.ctx, status, 3)
    
    // Process in parallel
    go s.logStatus(streams[0])
    go s.updateMetrics(streams[1])
    go s.renderDisplay(streams[2])
}
```

### Or-Channel Pattern

Combine multiple done channels into a single channel that closes when any input closes:

```go
// or.go
package statusline

func or(channels ...<-chan struct{}) <-chan struct{} {
    switch len(channels) {
    case 0:
        return nil
    case 1:
        return channels[0]
    }
    
    orDone := make(chan struct{})
    go func() {
        defer close(orDone)
        
        switch len(channels) {
        case 2:
            select {
            case <-channels[0]:
            case <-channels[1]:
            }
        default:
            select {
            case <-channels[0]:
            case <-channels[1]:
            case <-channels[2]:
            case <-or(append(channels[3:], orDone)...):
            }
        }
    }()
    
    return orDone
}

// Usage: Wait for first completion signal
func (s *StatusLine) waitForFirst(timeout <-chan struct{}, shutdown <-chan struct{}) {
    select {
    case <-or(timeout, shutdown, s.ctx.Done()):
        s.cleanup()
    }
}
```

### Bridge-Channel Pattern

Convert a channel of channels into a single channel:

```go
// bridge.go
package statusline

func bridge(ctx context.Context, chanStream <-chan <-chan interface{}) <-chan interface{} {
    valStream := make(chan interface{})
    
    go func() {
        defer close(valStream)
        for {
            var stream <-chan interface{}
            select {
            case maybeStream, ok := <-chanStream:
                if !ok {
                    return
                }
                stream = maybeStream
            case <-ctx.Done():
                return
            }
            
            for val := range orDone(ctx.Done(), stream) {
                select {
                case valStream <- val:
                case <-ctx.Done():
                }
            }
        }
    }()
    
    return valStream
}
```

## Worker Pools for Parallel Widget Rendering

### Basic Worker Pool

```go
// worker_pool.go
package statusline

import (
    "context"
    "sync"
    "time"
)

type RenderJob struct {
    Widget   Widget
    Priority int
    Result   chan<- RenderResult
}

type RenderResult struct {
    WidgetID string
    Content  string
    Error    error
    Duration time.Duration
}

type WorkerPool struct {
    workers    int
    jobs       chan RenderJob
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
    metrics    *PoolMetrics
}

type PoolMetrics struct {
    jobsProcessed   uint64
    jobsFailed      uint64
    avgRenderTime   time.Duration
    maxRenderTime   time.Duration
}

// NewWorkerPool creates a pool optimized for widget rendering
func NewWorkerPool(workers int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan RenderJob, workers*2), // 2x buffering
        ctx:     ctx,
        cancel:  cancel,
        metrics: &PoolMetrics{},
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(i)
    }
}

func (p *WorkerPool) worker(id int) {
    defer p.wg.Done()
    
    for {
        select {
        case <-p.ctx.Done():
            return
        case job, ok := <-p.jobs:
            if !ok {
                return
            }
            p.processJob(job)
        }
    }
}

func (p *WorkerPool) processJob(job RenderJob) {
    start := time.Now()
    
    // Create timeout context for individual widget
    ctx, cancel := context.WithTimeout(p.ctx, 5*time.Millisecond)
    defer cancel()
    
    // Render widget
    content, err := job.Widget.Render(ctx)
    duration := time.Since(start)
    
    // Update metrics
    atomic.AddUint64(&p.metrics.jobsProcessed, 1)
    if err != nil {
        atomic.AddUint64(&p.metrics.jobsFailed, 1)
    }
    
    // Update max render time atomically
    for {
        oldMax := atomic.LoadInt64((*int64)(&p.metrics.maxRenderTime))
        if duration.Nanoseconds() <= oldMax {
            break
        }
        if atomic.CompareAndSwapInt64((*int64)(&p.metrics.maxRenderTime), oldMax, duration.Nanoseconds()) {
            break
        }
    }
    
    result := RenderResult{
        WidgetID: job.Widget.ID(),
        Content:  content,
        Error:    err,
        Duration: duration,
    }
    
    select {
    case job.Result <- result:
    case <-p.ctx.Done():
    }
}

func (p *WorkerPool) Submit(job RenderJob) error {
    select {
    case p.jobs <- job:
        return nil
    case <-p.ctx.Done():
        return p.ctx.Err()
    default:
        return ErrPoolFull
    }
}

func (p *WorkerPool) SubmitBatch(widgets []Widget) ([]RenderResult, error) {
    resultChan := make(chan RenderResult, len(widgets))
    
    // Submit all jobs
    for i, widget := range widgets {
        job := RenderJob{
            Widget:   widget,
            Priority: i, // Maintain order
            Result:   resultChan,
        }
        
        if err := p.Submit(job); err != nil {
            return nil, err
        }
    }
    
    // Collect results
    results := make([]RenderResult, 0, len(widgets))
    for i := 0; i < len(widgets); i++ {
        select {
        case result := <-resultChan:
            results = append(results, result)
        case <-p.ctx.Done():
            return nil, p.ctx.Err()
        }
    }
    
    return results, nil
}

func (p *WorkerPool) Stop() {
    p.cancel()
    close(p.jobs)
    p.wg.Wait()
}

func (p *WorkerPool) Metrics() PoolMetrics {
    return PoolMetrics{
        jobsProcessed: atomic.LoadUint64(&p.metrics.jobsProcessed),
        jobsFailed:    atomic.LoadUint64(&p.metrics.jobsFailed),
        maxRenderTime: time.Duration(atomic.LoadInt64((*int64)(&p.metrics.maxRenderTime))),
    }
}
```

### Semaphore-Based Pool with Priority

```go
// semaphore_pool.go
package statusline

import (
    "container/heap"
    "context"
    "golang.org/x/sync/semaphore"
    "sync"
)

type PriorityPool struct {
    sem      *semaphore.Weighted
    queue    PriorityQueue
    mu       sync.Mutex
    cond     *sync.Cond
}

type PriorityJob struct {
    Widget   Widget
    Priority int
    Index    int // Used by heap
}

type PriorityQueue []*PriorityJob

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Priority > pq[j].Priority }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].Index = i; pq[j].Index = j }

func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    item := x.(*PriorityJob)
    item.Index = n
    *pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    old[n-1] = nil
    item.Index = -1
    *pq = old[0 : n-1]
    return item
}

func NewPriorityPool(maxConcurrent int64) *PriorityPool {
    pp := &PriorityPool{
        sem:   semaphore.NewWeighted(maxConcurrent),
        queue: make(PriorityQueue, 0),
    }
    pp.cond = sync.NewCond(&pp.mu)
    heap.Init(&pp.queue)
    return pp
}

func (pp *PriorityPool) Submit(ctx context.Context, widget Widget, priority int) error {
    pp.mu.Lock()
    heap.Push(&pp.queue, &PriorityJob{
        Widget:   widget,
        Priority: priority,
    })
    pp.mu.Unlock()
    pp.cond.Signal()
    
    return nil
}

func (pp *PriorityPool) Start(ctx context.Context) {
    for {
        pp.mu.Lock()
        for pp.queue.Len() == 0 {
            pp.cond.Wait()
            if ctx.Err() != nil {
                pp.mu.Unlock()
                return
            }
        }
        
        job := heap.Pop(&pp.queue).(*PriorityJob)
        pp.mu.Unlock()
        
        // Process with semaphore
        go func(j *PriorityJob) {
            if err := pp.sem.Acquire(ctx, 1); err != nil {
                return
            }
            defer pp.sem.Release(1)
            
            j.Widget.Render(ctx)
        }(job)
    }
}
```

## Context Propagation and Cancellation

### Context Hierarchy for Statusline

```go
// context_manager.go
package statusline

import (
    "context"
    "time"
)

type ContextManager struct {
    root     context.Context
    cancel   context.CancelFunc
    children map[string]context.CancelFunc
    mu       sync.RWMutex
}

func NewContextManager() *ContextManager {
    ctx, cancel := context.WithCancel(context.Background())
    return &ContextManager{
        root:     ctx,
        cancel:   cancel,
        children: make(map[string]context.CancelFunc),
    }
}

// CreateRenderContext creates a context for a single render cycle
func (cm *ContextManager) CreateRenderContext(timeout time.Duration) (context.Context, context.CancelFunc) {
    return context.WithTimeout(cm.root, timeout)
}

// CreateWidgetContext creates a context for individual widget rendering
func (cm *ContextManager) CreateWidgetContext(parent context.Context, widgetID string, timeout time.Duration) context.Context {
    // Add widget ID to context for tracing
    ctx := context.WithValue(parent, "widgetID", widgetID)
    
    // Apply timeout
    ctx, cancel := context.WithTimeout(ctx, timeout)
    
    // Track for cleanup
    cm.mu.Lock()
    cm.children[widgetID] = cancel
    cm.mu.Unlock()
    
    return ctx
}

// CancelWidget cancels a specific widget's context
func (cm *ContextManager) CancelWidget(widgetID string) {
    cm.mu.Lock()
    if cancel, ok := cm.children[widgetID]; ok {
        cancel()
        delete(cm.children, widgetID)
    }
    cm.mu.Unlock()
}

// Shutdown cancels all contexts
func (cm *ContextManager) Shutdown() {
    cm.cancel()
    
    // Cancel all children
    cm.mu.Lock()
    for _, cancel := range cm.children {
        cancel()
    }
    cm.children = make(map[string]context.CancelFunc)
    cm.mu.Unlock()
}

// Usage in statusline
func (s *StatusLine) renderCycle() error {
    // Create context for entire render cycle (10ms budget)
    ctx, cancel := s.ctxManager.CreateRenderContext(10 * time.Millisecond)
    defer cancel()
    
    // Track render time
    start := time.Now()
    
    // Render widgets in parallel
    results := make(chan RenderResult, len(s.widgets))
    
    for _, widget := range s.widgets {
        widget := widget // Capture for goroutine
        
        go func() {
            // Create widget-specific context (5ms budget)
            wctx := s.ctxManager.CreateWidgetContext(ctx, widget.ID(), 5*time.Millisecond)
            
            content, err := widget.Render(wctx)
            results <- RenderResult{
                WidgetID: widget.ID(),
                Content:  content,
                Error:    err,
            }
        }()
    }
    
    // Collect results
    var outputs []string
    for i := 0; i < len(s.widgets); i++ {
        select {
        case result := <-results:
            if result.Error != nil {
                outputs = append(outputs, fmt.Sprintf("[%s: error]", result.WidgetID))
            } else {
                outputs = append(outputs, result.Content)
            }
        case <-ctx.Done():
            return fmt.Errorf("render timeout exceeded: %v", time.Since(start))
        }
    }
    
    s.display(strings.Join(outputs, " | "))
    return nil
}
```

### Context Values for Request Scoping

```go
// context_values.go
package statusline

type contextKey int

const (
    sessionIDKey contextKey = iota
    requestIDKey
    widgetIDKey
    renderStartKey
)

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
    return context.WithValue(ctx, sessionIDKey, sessionID)
}

// SessionIDFromContext retrieves session ID from context
func SessionIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(sessionIDKey).(string)
    return id, ok
}

// WithRequestID adds request ID for tracing
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, requestIDKey, requestID)
}

// WithRenderMetadata adds render metadata
func WithRenderMetadata(ctx context.Context, widgetID string) context.Context {
    ctx = context.WithValue(ctx, widgetIDKey, widgetID)
    ctx = context.WithValue(ctx, renderStartKey, time.Now())
    return ctx
}

// GetRenderDuration calculates render duration from context
func GetRenderDuration(ctx context.Context) time.Duration {
    if start, ok := ctx.Value(renderStartKey).(time.Time); ok {
        return time.Since(start)
    }
    return 0
}

// Example widget that uses context values
type TracedWidget struct {
    BaseWidget
}

func (w *TracedWidget) Render(ctx context.Context) (string, error) {
    // Extract metadata from context
    widgetID, _ := ctx.Value(widgetIDKey).(string)
    sessionID, _ := SessionIDFromContext(ctx)
    
    // Log render start
    log.Printf("Rendering widget %s for session %s", widgetID, sessionID)
    
    // Perform render
    result, err := w.BaseWidget.Render(ctx)
    
    // Log render completion with duration
    duration := GetRenderDuration(ctx)
    log.Printf("Widget %s rendered in %v", widgetID, duration)
    
    return result, err
}
```

## Rate Limiting Patterns

### Token Bucket Rate Limiter

```go
// rate_limiter.go
package statusline

import (
    "context"
    "golang.org/x/time/rate"
    "sync"
    "time"
)

type RateLimitedStatusLine struct {
    limiter      *rate.Limiter
    widgets      []Widget
    burstLimiter *BurstLimiter
}

// NewRateLimitedStatusLine creates a statusline with rate limiting
func NewRateLimitedStatusLine(rps int, burst int) *RateLimitedStatusLine {
    return &RateLimitedStatusLine{
        limiter:      rate.NewLimiter(rate.Limit(rps), burst),
        burstLimiter: NewBurstLimiter(burst, time.Second),
    }
}

func (rl *RateLimitedStatusLine) Update(ctx context.Context) error {
    // Wait for rate limit token
    if err := rl.limiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limit wait: %w", err)
    }
    
    // Check burst limit
    if !rl.burstLimiter.Allow() {
        return ErrBurstLimitExceeded
    }
    
    return rl.render(ctx)
}

// BurstLimiter prevents sudden spikes
type BurstLimiter struct {
    mu       sync.Mutex
    tokens   int
    capacity int
    refill   time.Duration
    lastRefill time.Time
}

func NewBurstLimiter(capacity int, refillInterval time.Duration) *BurstLimiter {
    return &BurstLimiter{
        tokens:     capacity,
        capacity:   capacity,
        refill:     refillInterval,
        lastRefill: time.Now(),
    }
}

func (bl *BurstLimiter) Allow() bool {
    bl.mu.Lock()
    defer bl.mu.Unlock()
    
    // Refill tokens based on elapsed time
    now := time.Now()
    elapsed := now.Sub(bl.lastRefill)
    tokensToAdd := int(elapsed / bl.refill)
    
    if tokensToAdd > 0 {
        bl.tokens = min(bl.capacity, bl.tokens+tokensToAdd)
        bl.lastRefill = now
    }
    
    if bl.tokens > 0 {
        bl.tokens--
        return true
    }
    
    return false
}

// Per-widget rate limiting
type WidgetRateLimiter struct {
    limiters sync.Map // map[string]*rate.Limiter
    defaults rate.Limit
    burst    int
}

func NewWidgetRateLimiter(defaultRPS float64, burst int) *WidgetRateLimiter {
    return &WidgetRateLimiter{
        defaults: rate.Limit(defaultRPS),
        burst:    burst,
    }
}

func (wrl *WidgetRateLimiter) GetLimiter(widgetID string) *rate.Limiter {
    if limiter, ok := wrl.limiters.Load(widgetID); ok {
        return limiter.(*rate.Limiter)
    }
    
    // Create new limiter
    limiter := rate.NewLimiter(wrl.defaults, wrl.burst)
    actual, _ := wrl.limiters.LoadOrStore(widgetID, limiter)
    return actual.(*rate.Limiter)
}

func (wrl *WidgetRateLimiter) SetWidgetRate(widgetID string, rps float64) {
    limiter := rate.NewLimiter(rate.Limit(rps), wrl.burst)
    wrl.limiters.Store(widgetID, limiter)
}
```

### Adaptive Rate Limiting

```go
// adaptive_rate.go
package statusline

import (
    "context"
    "sync"
    "sync/atomic"
    "time"
)

type AdaptiveRateLimiter struct {
    mu            sync.RWMutex
    currentRate   rate.Limit
    minRate       rate.Limit
    maxRate       rate.Limit
    window        *MetricsWindow
    adjustPeriod  time.Duration
    lastAdjust    time.Time
}

type MetricsWindow struct {
    successes uint64
    failures  uint64
    latencies []time.Duration
    mu        sync.Mutex
}

func NewAdaptiveRateLimiter(minRPS, maxRPS float64) *AdaptiveRateLimiter {
    return &AdaptiveRateLimiter{
        currentRate:  rate.Limit((minRPS + maxRPS) / 2),
        minRate:      rate.Limit(minRPS),
        maxRate:      rate.Limit(maxRPS),
        window:       &MetricsWindow{latencies: make([]time.Duration, 0, 100)},
        adjustPeriod: 10 * time.Second,
        lastAdjust:   time.Now(),
    }
}

func (a *AdaptiveRateLimiter) RecordSuccess(latency time.Duration) {
    atomic.AddUint64(&a.window.successes, 1)
    
    a.window.mu.Lock()
    a.window.latencies = append(a.window.latencies, latency)
    if len(a.window.latencies) > 100 {
        a.window.latencies = a.window.latencies[1:]
    }
    a.window.mu.Unlock()
    
    a.maybeAdjust()
}

func (a *AdaptiveRateLimiter) RecordFailure() {
    atomic.AddUint64(&a.window.failures, 1)
    a.maybeAdjust()
}

func (a *AdaptiveRateLimiter) maybeAdjust() {
    now := time.Now()
    if now.Sub(a.lastAdjust) < a.adjustPeriod {
        return
    }
    
    a.mu.Lock()
    defer a.mu.Unlock()
    
    // Check if we should adjust again
    if now.Sub(a.lastAdjust) < a.adjustPeriod {
        return
    }
    
    successes := atomic.LoadUint64(&a.window.successes)
    failures := atomic.LoadUint64(&a.window.failures)
    total := successes + failures
    
    if total == 0 {
        return
    }
    
    errorRate := float64(failures) / float64(total)
    p95Latency := a.calculateP95Latency()
    
    // Adjust rate based on error rate and latency
    if errorRate > 0.05 || p95Latency > 10*time.Millisecond {
        // Reduce rate by 20%
        a.currentRate = rate.Limit(float64(a.currentRate) * 0.8)
        if a.currentRate < a.minRate {
            a.currentRate = a.minRate
        }
    } else if errorRate < 0.01 && p95Latency < 5*time.Millisecond {
        // Increase rate by 10%
        a.currentRate = rate.Limit(float64(a.currentRate) * 1.1)
        if a.currentRate > a.maxRate {
            a.currentRate = a.maxRate
        }
    }
    
    // Reset window
    atomic.StoreUint64(&a.window.successes, 0)
    atomic.StoreUint64(&a.window.failures, 0)
    a.window.mu.Lock()
    a.window.latencies = a.window.latencies[:0]
    a.window.mu.Unlock()
    
    a.lastAdjust = now
}

func (a *AdaptiveRateLimiter) calculateP95Latency() time.Duration {
    a.window.mu.Lock()
    defer a.window.mu.Unlock()
    
    if len(a.window.latencies) == 0 {
        return 0
    }
    
    // Simple P95 calculation (could be optimized)
    sorted := make([]time.Duration, len(a.window.latencies))
    copy(sorted, a.window.latencies)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i] < sorted[j]
    })
    
    idx := int(float64(len(sorted)) * 0.95)
    return sorted[idx]
}

func (a *AdaptiveRateLimiter) Wait(ctx context.Context) error {
    a.mu.RLock()
    limiter := rate.NewLimiter(a.currentRate, 1)
    a.mu.RUnlock()
    
    return limiter.Wait(ctx)
}
```

## Synchronization with sync Package

### Mutex vs RWMutex

```go
// sync_primitives.go
package statusline

import (
    "sync"
    "sync/atomic"
)

// StatusCache with RWMutex for read-heavy workloads
type StatusCache struct {
    mu      sync.RWMutex
    data    map[string]interface{}
    version uint64 // For cache invalidation
}

func NewStatusCache() *StatusCache {
    return &StatusCache{
        data: make(map[string]interface{}),
    }
}

// Get allows multiple concurrent readers
func (c *StatusCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    val, ok := c.data[key]
    return val, ok
}

// GetVersion returns current cache version without blocking writers
func (c *StatusCache) GetVersion() uint64 {
    return atomic.LoadUint64(&c.version)
}

// Set requires exclusive access
func (c *StatusCache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
    atomic.AddUint64(&c.version, 1)
}

// SetBatch updates multiple values atomically
func (c *StatusCache) SetBatch(updates map[string]interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    for k, v := range updates {
        c.data[k] = v
    }
    atomic.AddUint64(&c.version, 1)
}

// GetAll returns a snapshot of all data
func (c *StatusCache) GetAll() map[string]interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    // Create copy to avoid holding lock during processing
    snapshot := make(map[string]interface{}, len(c.data))
    for k, v := range c.data {
        snapshot[k] = v
    }
    
    return snapshot
}
```

### WaitGroup for Complex Synchronization

```go
// waitgroup_patterns.go
package statusline

import (
    "context"
    "errors"
    "sync"
    "time"
)

// ParallelRenderer renders widgets in parallel with proper synchronization
type ParallelRenderer struct {
    timeout time.Duration
}

func (pr *ParallelRenderer) RenderAll(ctx context.Context, widgets []Widget) ([]string, error) {
    results := make([]string, len(widgets))
    errors := make([]error, len(widgets))
    var wg sync.WaitGroup
    
    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, pr.timeout)
    defer cancel()
    
    // Launch goroutine for each widget
    for i, widget := range widgets {
        wg.Add(1)
        go func(idx int, w Widget) {
            defer wg.Done()
            
            // Render with panic recovery
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        errors[idx] = fmt.Errorf("panic in widget %s: %v", w.ID(), r)
                    }
                }()
                
                results[idx], errors[idx] = w.Render(ctx)
            }()
        }(i, widget)
    }
    
    // Wait for completion or timeout
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        // All completed
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    // Check for errors
    for i, err := range errors {
        if err != nil {
            return nil, fmt.Errorf("widget %d failed: %w", i, err)
        }
    }
    
    return results, nil
}

// ErrGroup-like pattern for better error handling
type WidgetGroup struct {
    wg      sync.WaitGroup
    errOnce sync.Once
    err     error
    ctx     context.Context
    cancel  context.CancelFunc
}

func NewWidgetGroup(ctx context.Context) *WidgetGroup {
    ctx, cancel := context.WithCancel(ctx)
    return &WidgetGroup{
        ctx:    ctx,
        cancel: cancel,
    }
}

func (g *WidgetGroup) Go(f func() error) {
    g.wg.Add(1)
    
    go func() {
        defer g.wg.Done()
        
        if err := f(); err != nil {
            g.errOnce.Do(func() {
                g.err = err
                g.cancel()
            })
        }
    }()
}

func (g *WidgetGroup) Wait() error {
    g.wg.Wait()
    g.cancel()
    return g.err
}

// Usage
func (s *StatusLine) renderWithGroup(widgets []Widget) error {
    g := NewWidgetGroup(s.ctx)
    results := make([]string, len(widgets))
    
    for i, widget := range widgets {
        i, widget := i, widget // Capture for closure
        g.Go(func() error {
            result, err := widget.Render(g.ctx)
            if err != nil {
                return err
            }
            results[i] = result
            return nil
        })
    }
    
    if err := g.Wait(); err != nil {
        return err
    }
    
    s.display(strings.Join(results, " | "))
    return nil
}
```

### Once for One-Time Initialization

```go
// once_patterns.go
package statusline

import (
    "sync"
    "sync/atomic"
)

// ThemeManager with lazy initialization
type ThemeManager struct {
    once      sync.Once
    themes    map[string]Theme
    current   atomic.Value // stores Theme
    initError error
}

func (tm *ThemeManager) initialize() {
    tm.themes = make(map[string]Theme)
    
    // Load built-in themes
    tm.themes["classic"] = NewClassicTheme()
    tm.themes["powerline"] = NewPowerlineTheme()
    tm.themes["capsule"] = NewCapsuleTheme()
    
    // Load custom themes
    if err := tm.loadCustomThemes(); err != nil {
        tm.initError = err
        return
    }
    
    // Set default theme
    if theme, ok := tm.themes["classic"]; ok {
        tm.current.Store(theme)
    }
}

func (tm *ThemeManager) GetTheme(name string) (Theme, error) {
    tm.once.Do(tm.initialize)
    
    if tm.initError != nil {
        return nil, tm.initError
    }
    
    theme, ok := tm.themes[name]
    if !ok {
        return nil, ErrThemeNotFound
    }
    
    return theme, nil
}

func (tm *ThemeManager) GetCurrent() Theme {
    tm.once.Do(tm.initialize)
    
    if theme := tm.current.Load(); theme != nil {
        return theme.(Theme)
    }
    
    return NewClassicTheme() // Fallback
}

// Terminal capabilities detection with once
var (
    termCapsOnce sync.Once
    termCaps     *TerminalCapabilities
)

func GetTerminalCapabilities() *TerminalCapabilities {
    termCapsOnce.Do(func() {
        termCaps = detectCapabilities()
    })
    return termCaps
}
```

### Cond for Complex Coordination

```go
// cond_patterns.go
package statusline

import (
    "sync"
    "time"
)

// EventBroadcaster using Cond for efficient event distribution
type EventBroadcaster struct {
    mu        sync.Mutex
    cond      *sync.Cond
    events    []Event
    maxEvents int
}

func NewEventBroadcaster(maxEvents int) *EventBroadcaster {
    eb := &EventBroadcaster{
        maxEvents: maxEvents,
        events:    make([]Event, 0, maxEvents),
    }
    eb.cond = sync.NewCond(&eb.mu)
    return eb
}

func (eb *EventBroadcaster) Broadcast(event Event) {
    eb.mu.Lock()
    eb.events = append(eb.events, event)
    if len(eb.events) > eb.maxEvents {
        eb.events = eb.events[1:]
    }
    eb.mu.Unlock()
    
    eb.cond.Broadcast()
}

func (eb *EventBroadcaster) Wait() Event {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    eb.cond.Wait()
    if len(eb.events) > 0 {
        return eb.events[len(eb.events)-1]
    }
    return Event{}
}

func (eb *EventBroadcaster) WaitTimeout(timeout time.Duration) (Event, bool) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    done := make(chan struct{})
    timer := time.AfterFunc(timeout, func() {
        close(done)
        eb.cond.Broadcast()
    })
    defer timer.Stop()
    
    eb.cond.Wait()
    
    select {
    case <-done:
        return Event{}, false
    default:
        if len(eb.events) > 0 {
            return eb.events[len(eb.events)-1], true
        }
        return Event{}, false
    }
}

// BatchProcessor using Cond for batch triggers
type BatchProcessor struct {
    mu        sync.Mutex
    cond      *sync.Cond
    items     []interface{}
    batchSize int
    processor func([]interface{}) error
}

func NewBatchProcessor(batchSize int, processor func([]interface{}) error) *BatchProcessor {
    bp := &BatchProcessor{
        batchSize: batchSize,
        processor: processor,
        items:     make([]interface{}, 0, batchSize),
    }
    bp.cond = sync.NewCond(&bp.mu)
    go bp.processLoop()
    return bp
}

func (bp *BatchProcessor) Add(item interface{}) {
    bp.mu.Lock()
    bp.items = append(bp.items, item)
    if len(bp.items) >= bp.batchSize {
        bp.cond.Signal()
    }
    bp.mu.Unlock()
}

func (bp *BatchProcessor) processLoop() {
    for {
        bp.mu.Lock()
        
        // Wait for batch to fill
        for len(bp.items) < bp.batchSize {
            bp.cond.Wait()
        }
        
        // Process batch
        batch := bp.items
        bp.items = make([]interface{}, 0, bp.batchSize)
        bp.mu.Unlock()
        
        if err := bp.processor(batch); err != nil {
            // Handle error
        }
    }
}
```

### sync.Map for Concurrent Access

```go
// sync_map_patterns.go
package statusline

import (
    "sync"
    "sync/atomic"
    "time"
)

// MetricsCollector using sync.Map for lock-free reads
type MetricsCollector struct {
    counters  sync.Map // map[string]*int64
    gauges    sync.Map // map[string]*float64
    histograms sync.Map // map[string]*Histogram
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{}
}

func (m *MetricsCollector) IncrementCounter(name string, delta int64) {
    actual, _ := m.counters.LoadOrStore(name, new(int64))
    atomic.AddInt64(actual.(*int64), delta)
}

func (m *MetricsCollector) SetGauge(name string, value float64) {
    actual, _ := m.gauges.LoadOrStore(name, new(float64))
    atomic.StoreUint64((*uint64)(actual.(*float64)), math.Float64bits(value))
}

func (m *MetricsCollector) RecordDuration(name string, duration time.Duration) {
    actual, loaded := m.histograms.LoadOrStore(name, NewHistogram())
    histogram := actual.(*Histogram)
    histogram.Record(duration)
}

func (m *MetricsCollector) GetSnapshot() map[string]interface{} {
    snapshot := make(map[string]interface{})
    
    // Collect counters
    m.counters.Range(func(key, value interface{}) bool {
        snapshot[key.(string)] = atomic.LoadInt64(value.(*int64))
        return true
    })
    
    // Collect gauges
    m.gauges.Range(func(key, value interface{}) bool {
        bits := atomic.LoadUint64((*uint64)(value.(*float64)))
        snapshot[key.(string)] = math.Float64frombits(bits)
        return true
    })
    
    return snapshot
}

// WidgetRegistry with concurrent registration
type WidgetRegistry struct {
    widgets sync.Map // map[string]WidgetFactory
}

func (wr *WidgetRegistry) Register(name string, factory WidgetFactory) {
    wr.widgets.Store(name, factory)
}

func (wr *WidgetRegistry) Create(name string, config interface{}) (Widget, error) {
    factory, ok := wr.widgets.Load(name)
    if !ok {
        return nil, ErrWidgetNotFound
    }
    
    return factory.(WidgetFactory)(config)
}

func (wr *WidgetRegistry) List() []string {
    var names []string
    wr.widgets.Range(func(key, _ interface{}) bool {
        names = append(names, key.(string))
        return true
    })
    return names
}
```

## Preventing Goroutine Leaks

### Always Use Context

```go
// leak_prevention.go
package statusline

import (
    "context"
    "sync"
    "time"
)

// SafeWorker demonstrates leak-proof goroutine patterns
type SafeWorker struct {
    ctx    context.Context
    cancel context.CancelFunc
    done   chan struct{}
    wg     sync.WaitGroup
}

func NewSafeWorker() *SafeWorker {
    ctx, cancel := context.WithCancel(context.Background())
    return &SafeWorker{
        ctx:    ctx,
        cancel: cancel,
        done:   make(chan struct{}),
    }
}

// Start launches a leak-proof worker
func (sw *SafeWorker) Start() {
    sw.wg.Add(1)
    go func() {
        defer sw.wg.Done()
        defer close(sw.done)
        
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-sw.ctx.Done():
                // Clean exit on context cancellation
                return
            case <-ticker.C:
                // Do work, but check context during long operations
                if err := sw.doWork(sw.ctx); err != nil {
                    if err == context.Canceled {
                        return
                    }
                    // Handle other errors
                }
            }
        }
    }()
}

func (sw *SafeWorker) doWork(ctx context.Context) error {
    // Simulate work that respects context
    workDone := make(chan struct{})
    
    go func() {
        // Actual work here
        time.Sleep(100 * time.Millisecond)
        close(workDone)
    }()
    
    select {
    case <-workDone:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (sw *SafeWorker) Stop() error {
    sw.cancel()
    
    // Wait with timeout
    done := make(chan struct{})
    go func() {
        sw.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(5 * time.Second):
        return ErrShutdownTimeout
    }
}

// SafeChannelWorker shows channel ownership patterns
type SafeChannelWorker struct {
    input  chan Task
    output chan Result
    done   chan struct{}
}

func NewSafeChannelWorker() *SafeChannelWorker {
    return &SafeChannelWorker{
        input:  make(chan Task),
        output: make(chan Result),
        done:   make(chan struct{}),
    }
}

func (scw *SafeChannelWorker) Start(ctx context.Context) {
    go func() {
        defer close(scw.output) // Producer closes channel
        
        for {
            select {
            case task, ok := <-scw.input:
                if !ok {
                    return // Input closed, exit gracefully
                }
                
                result := scw.process(task)
                
                select {
                case scw.output <- result:
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()
}

func (scw *SafeChannelWorker) Stop() {
    close(scw.input) // Signal worker to stop
    <-scw.done       // Wait for completion
}
```

### Leak Detection in Tests

```go
// leak_test.go
package statusline

import (
    "testing"
    "time"
    "go.uber.org/goleak"
)

func TestNoGoroutineLeaks(t *testing.T) {
    defer goleak.VerifyNone(t)
    
    // Start statusline
    sl := NewStatusLine()
    ctx, cancel := context.WithCancel(context.Background())
    
    go sl.Run(ctx)
    
    // Let it run
    time.Sleep(100 * time.Millisecond)
    
    // Stop and verify cleanup
    cancel()
    time.Sleep(50 * time.Millisecond) // Grace period
}

// Custom leak detector
func TestCustomLeakDetection(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // Run test
    sl := NewStatusLine()
    sl.Start()
    sl.Stop()
    
    // Check after cleanup
    time.Sleep(100 * time.Millisecond)
    after := runtime.NumGoroutine()
    
    if after > before {
        t.Errorf("Goroutine leak detected: before=%d, after=%d", before, after)
        
        // Print stack traces for debugging
        buf := make([]byte, 1<<20)
        stackSize := runtime.Stack(buf, true)
        t.Logf("Goroutine stack traces:\n%s", buf[:stackSize])
    }
}

// Benchmark with leak detection
func BenchmarkStatusLineNoLeaks(b *testing.B) {
    for i := 0; i < b.N; i++ {
        sl := NewStatusLine()
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
        
        sl.RenderOnce(ctx)
        cancel()
        
        // Verify no accumulation
        if i%100 == 0 && runtime.NumGoroutine() > 100 {
            b.Fatalf("Goroutine leak: %d goroutines at iteration %d", 
                runtime.NumGoroutine(), i)
        }
    }
}
```

### Resource Cleanup Patterns

```go
// cleanup_patterns.go
package statusline

import (
    "context"
    "sync"
    "time"
)

// ResourceManager ensures proper cleanup of all resources
type ResourceManager struct {
    mu        sync.Mutex
    resources []Resource
    closed    bool
}

type Resource interface {
    Close() error
}

func (rm *ResourceManager) Add(r Resource) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    if rm.closed {
        return ErrManagerClosed
    }
    
    rm.resources = append(rm.resources, r)
    return nil
}

func (rm *ResourceManager) Close() error {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    if rm.closed {
        return nil
    }
    
    rm.closed = true
    
    var errs []error
    // Close in reverse order (LIFO)
    for i := len(rm.resources) - 1; i >= 0; i-- {
        if err := rm.resources[i].Close(); err != nil {
            errs = append(errs, err)
        }
    }
    
    if len(errs) > 0 {
        return MultiError(errs)
    }
    
    return nil
}

// CleanupStack for deterministic cleanup
type CleanupStack struct {
    mu       sync.Mutex
    cleanups []func() error
}

func (cs *CleanupStack) Push(cleanup func() error) {
    cs.mu.Lock()
    cs.cleanups = append(cs.cleanups, cleanup)
    cs.mu.Unlock()
}

func (cs *CleanupStack) Cleanup() error {
    cs.mu.Lock()
    defer cs.mu.Unlock()
    
    var errs []error
    
    // Execute in LIFO order
    for i := len(cs.cleanups) - 1; i >= 0; i-- {
        if err := cs.cleanups[i](); err != nil {
            errs = append(errs, err)
        }
    }
    
    cs.cleanups = nil
    
    if len(errs) > 0 {
        return MultiError(errs)
    }
    
    return nil
}

// Usage pattern
func (s *StatusLine) Initialize() error {
    cleanup := &CleanupStack{}
    
    // Initialize worker pool
    pool := NewWorkerPool(runtime.NumCPU())
    pool.Start()
    cleanup.Push(pool.Stop)
    
    // Initialize rate limiter
    limiter := NewRateLimiter(100)
    cleanup.Push(limiter.Stop)
    
    // Initialize metrics
    metrics := NewMetricsCollector()
    cleanup.Push(metrics.Close)
    
    // If any initialization fails, cleanup everything
    if err := s.connectToBackend(); err != nil {
        cleanup.Cleanup()
        return err
    }
    
    s.cleanup = cleanup
    return nil
}
```

## Backpressure and Flow Control

### Buffered Channel Backpressure

```go
// backpressure.go
package statusline

import (
    "context"
    "errors"
    "sync"
    "time"
)

// BufferedPipeline with backpressure control
type BufferedPipeline struct {
    stages    []Stage
    buffers   []chan Message
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    metrics   *PipelineMetrics
}

type Stage interface {
    Process(context.Context, Message) (Message, error)
}

type Message struct {
    ID        string
    Data      interface{}
    Timestamp time.Time
    Priority  int
}

type PipelineMetrics struct {
    processed   uint64
    dropped     uint64
    queueDepth  []int
    processingTime []time.Duration
}

func NewBufferedPipeline(bufferSize int, stages ...Stage) *BufferedPipeline {
    ctx, cancel := context.WithCancel(context.Background())
    
    buffers := make([]chan Message, len(stages)+1)
    for i := range buffers {
        buffers[i] = make(chan Message, bufferSize)
    }
    
    return &BufferedPipeline{
        stages:  stages,
        buffers: buffers,
        ctx:     ctx,
        cancel:  cancel,
        metrics: &PipelineMetrics{
            queueDepth:     make([]int, len(stages)),
            processingTime: make([]time.Duration, len(stages)),
        },
    }
}

func (bp *BufferedPipeline) Start() {
    for i, stage := range bp.stages {
        bp.wg.Add(1)
        go bp.runStage(i, stage)
    }
}

func (bp *BufferedPipeline) runStage(index int, stage Stage) {
    defer bp.wg.Done()
    
    input := bp.buffers[index]
    output := bp.buffers[index+1]
    
    for {
        select {
        case msg, ok := <-input:
            if !ok {
                close(output)
                return
            }
            
            start := time.Now()
            processed, err := stage.Process(bp.ctx, msg)
            duration := time.Since(start)
            
            // Update metrics
            atomic.AddUint64(&bp.metrics.processed, 1)
            bp.metrics.processingTime[index] = duration
            
            if err != nil {
                // Handle error - could send to error channel
                continue
            }
            
            // Apply backpressure
            select {
            case output <- processed:
                // Success
            case <-time.After(100 * time.Millisecond):
                // Timeout - downstream is slow
                atomic.AddUint64(&bp.metrics.dropped, 1)
            case <-bp.ctx.Done():
                return
            }
            
        case <-bp.ctx.Done():
            return
        }
    }
}

func (bp *BufferedPipeline) Submit(msg Message) error {
    select {
    case bp.buffers[0] <- msg:
        return nil
    case <-bp.ctx.Done():
        return bp.ctx.Err()
    default:
        // Buffer full - apply backpressure
        return ErrPipelineFull
    }
}

func (bp *BufferedPipeline) SubmitWithTimeout(msg Message, timeout time.Duration) error {
    select {
    case bp.buffers[0] <- msg:
        return nil
    case <-time.After(timeout):
        return ErrSubmitTimeout
    case <-bp.ctx.Done():
        return bp.ctx.Err()
    }
}

func (bp *BufferedPipeline) Stop() {
    bp.cancel()
    close(bp.buffers[0])
    bp.wg.Wait()
}

// Usage: Widget rendering pipeline with backpressure
func createRenderPipeline() *BufferedPipeline {
    collectStage := StageFunc(func(ctx context.Context, msg Message) (Message, error) {
        widget := msg.Data.(Widget)
        status, err := widget.CollectStatus(ctx)
        if err != nil {
            return msg, err
        }
        msg.Data = status
        return msg, nil
    })
    
    formatStage := StageFunc(func(ctx context.Context, msg Message) (Message, error) {
        status := msg.Data.(Status)
        formatted := theme.Format(status)
        msg.Data = formatted
        return msg, nil
    })
    
    renderStage := StageFunc(func(ctx context.Context, msg Message) (Message, error) {
        formatted := msg.Data.(string)
        terminal.Render(formatted)
        return msg, nil
    })
    
    return NewBufferedPipeline(10, collectStage, formatStage, renderStage)
}
```

### Semaphore-Based Flow Control

```go
// flow_control.go
package statusline

import (
    "context"
    "golang.org/x/sync/errgroup"
    "golang.org/x/sync/semaphore"
    "sync"
    "time"
)

// FlowController manages concurrent operations with semaphore
type FlowController struct {
    sem          *semaphore.Weighted
    maxInFlight  int64
    timeout      time.Duration
    metrics      *FlowMetrics
}

type FlowMetrics struct {
    acquired    uint64
    released    uint64
    timeouts    uint64
    queueDepth  int64
}

func NewFlowController(maxConcurrent int64, timeout time.Duration) *FlowController {
    return &FlowController{
        sem:         semaphore.NewWeighted(maxConcurrent),
        maxInFlight: maxConcurrent,
        timeout:     timeout,
        metrics:     &FlowMetrics{},
    }
}

func (fc *FlowController) Process(ctx context.Context, items []interface{}, processor func(interface{}) error) error {
    g, ctx := errgroup.WithContext(ctx)
    
    for _, item := range items {
        item := item // Capture for closure
        
        g.Go(func() error {
            // Try to acquire with timeout
            acquireCtx, cancel := context.WithTimeout(ctx, fc.timeout)
            defer cancel()
            
            if err := fc.sem.Acquire(acquireCtx, 1); err != nil {
                atomic.AddUint64(&fc.metrics.timeouts, 1)
                return err
            }
            atomic.AddUint64(&fc.metrics.acquired, 1)
            
            defer func() {
                fc.sem.Release(1)
                atomic.AddUint64(&fc.metrics.released, 1)
            }()
            
            return processor(item)
        })
    }
    
    return g.Wait()
}

func (fc *FlowController) ProcessBatched(ctx context.Context, items []interface{}, batchSize int, processor func([]interface{}) error) error {
    // Process in batches to reduce overhead
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        
        // Acquire resources for batch
        if err := fc.sem.Acquire(ctx, int64(len(batch))); err != nil {
            return err
        }
        
        err := processor(batch)
        fc.sem.Release(int64(len(batch)))
        
        if err != nil {
            return err
        }
    }
    
    return nil
}

func (fc *FlowController) TryProcess(item interface{}, processor func(interface{}) error) error {
    // Non-blocking acquire
    if !fc.sem.TryAcquire(1) {
        return ErrNoCapacity
    }
    
    defer fc.sem.Release(1)
    return processor(item)
}

func (fc *FlowController) QueueDepth() int {
    // Estimate queue depth
    available := fc.maxInFlight - int64(fc.metrics.acquired-fc.metrics.released)
    return int(fc.maxInFlight - available)
}
```

### Dynamic Worker Scaling

```go
// dynamic_scaling.go
package statusline

import (
    "context"
    "math"
    "sync"
    "sync/atomic"
    "time"
)

// DynamicPool scales workers based on load
type DynamicPool struct {
    mu          sync.RWMutex
    minWorkers  int
    maxWorkers  int
    workers     []*Worker
    taskQueue   chan Task
    metrics     *LoadMetrics
    scaler      *AutoScaler
}

type LoadMetrics struct {
    tasksQueued     int64
    tasksProcessed  int64
    avgProcessTime  float64
    queueTime       float64
}

type AutoScaler struct {
    checkInterval   time.Duration
    scaleUpThreshold   float64
    scaleDownThreshold float64
    cooldownPeriod     time.Duration
    lastScaleTime      time.Time
}

func NewDynamicPool(min, max int) *DynamicPool {
    dp := &DynamicPool{
        minWorkers: min,
        maxWorkers: max,
        taskQueue:  make(chan Task, max*10),
        metrics:    &LoadMetrics{},
        scaler: &AutoScaler{
            checkInterval:      5 * time.Second,
            scaleUpThreshold:   0.8,
            scaleDownThreshold: 0.2,
            cooldownPeriod:     30 * time.Second,
        },
    }
    
    // Start with minimum workers
    for i := 0; i < min; i++ {
        dp.addWorker()
    }
    
    go dp.autoScale()
    
    return dp
}

func (dp *DynamicPool) addWorker() {
    w := &Worker{
        id:    len(dp.workers),
        tasks: dp.taskQueue,
        pool:  dp,
    }
    
    dp.workers = append(dp.workers, w)
    go w.run()
}

func (dp *DynamicPool) removeWorker() {
    if len(dp.workers) <= dp.minWorkers {
        return
    }
    
    // Signal last worker to stop
    w := dp.workers[len(dp.workers)-1]
    w.stop()
    dp.workers = dp.workers[:len(dp.workers)-1]
}

func (dp *DynamicPool) autoScale() {
    ticker := time.NewTicker(dp.scaler.checkInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        dp.evaluateScaling()
    }
}

func (dp *DynamicPool) evaluateScaling() {
    dp.mu.Lock()
    defer dp.mu.Unlock()
    
    // Check cooldown
    if time.Since(dp.scaler.lastScaleTime) < dp.scaler.cooldownPeriod {
        return
    }
    
    // Calculate metrics
    queueSize := len(dp.taskQueue)
    capacity := cap(dp.taskQueue)
    utilization := float64(queueSize) / float64(capacity)
    
    currentWorkers := len(dp.workers)
    
    // Scale up if high utilization
    if utilization > dp.scaler.scaleUpThreshold && currentWorkers < dp.maxWorkers {
        newWorkers := int(math.Min(
            float64(currentWorkers*2),
            float64(dp.maxWorkers),
        ))
        
        for i := currentWorkers; i < newWorkers; i++ {
            dp.addWorker()
        }
        
        dp.scaler.lastScaleTime = time.Now()
        log.Printf("Scaled up to %d workers (utilization: %.2f)", newWorkers, utilization)
        
    // Scale down if low utilization
    } else if utilization < dp.scaler.scaleDownThreshold && currentWorkers > dp.minWorkers {
        newWorkers := int(math.Max(
            float64(currentWorkers/2),
            float64(dp.minWorkers),
        ))
        
        for i := currentWorkers; i > newWorkers; i-- {
            dp.removeWorker()
        }
        
        dp.scaler.lastScaleTime = time.Now()
        log.Printf("Scaled down to %d workers (utilization: %.2f)", newWorkers, utilization)
    }
}

func (dp *DynamicPool) Submit(task Task) error {
    atomic.AddInt64(&dp.metrics.tasksQueued, 1)
    
    select {
    case dp.taskQueue <- task:
        return nil
    default:
        return ErrQueueFull
    }
}

func (dp *DynamicPool) Shutdown(ctx context.Context) error {
    dp.mu.Lock()
    defer dp.mu.Unlock()
    
    // Stop all workers
    for _, w := range dp.workers {
        w.stop()
    }
    
    // Wait for completion or timeout
    done := make(chan struct{})
    go func() {
        for _, w := range dp.workers {
            w.wait()
        }
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Real-Time Event Streaming Patterns

### Event Bus Pattern

```go
// event_bus.go
package statusline

import (
    "context"
    "sync"
    "time"
)

type EventType string

const (
    EventStatusUpdate EventType = "status_update"
    EventConfigChange EventType = "config_change"
    EventError        EventType = "error"
    EventShutdown     EventType = "shutdown"
)

type Event struct {
    Type      EventType
    Timestamp time.Time
    Data      interface{}
    Source    string
}

type EventBus struct {
    mu          sync.RWMutex
    subscribers map[EventType]map[string]chan Event
    bufferSize  int
    metrics     *EventMetrics
}

type EventMetrics struct {
    published  map[EventType]uint64
    delivered  map[EventType]uint64
    dropped    map[EventType]uint64
}

func NewEventBus(bufferSize int) *EventBus {
    return &EventBus{
        subscribers: make(map[EventType]map[string]chan Event),
        bufferSize:  bufferSize,
        metrics: &EventMetrics{
            published: make(map[EventType]uint64),
            delivered: make(map[EventType]uint64),
            dropped:   make(map[EventType]uint64),
        },
    }
}

func (eb *EventBus) Subscribe(subscriberID string, eventType EventType, buffer int) <-chan Event {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    if _, ok := eb.subscribers[eventType]; !ok {
        eb.subscribers[eventType] = make(map[string]chan Event)
    }
    
    ch := make(chan Event, buffer)
    eb.subscribers[eventType][subscriberID] = ch
    
    return ch
}

func (eb *EventBus) Unsubscribe(subscriberID string, eventType EventType) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    if subs, ok := eb.subscribers[eventType]; ok {
        if ch, ok := subs[subscriberID]; ok {
            close(ch)
            delete(subs, subscriberID)
        }
    }
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    subscribers := eb.subscribers[event.Type]
    eb.mu.RUnlock()
    
    atomic.AddUint64(&eb.metrics.published[event.Type], 1)
    
    // Fan out to all subscribers
    var wg sync.WaitGroup
    for id, ch := range subscribers {
        wg.Add(1)
        go func(subscriberID string, channel chan Event) {
            defer wg.Done()
            
            select {
            case channel <- event:
                atomic.AddUint64(&eb.metrics.delivered[event.Type], 1)
            default:
                // Channel full, drop event
                atomic.AddUint64(&eb.metrics.dropped[event.Type], 1)
                log.Printf("Dropped event for subscriber %s: buffer full", subscriberID)
            }
        }(id, ch)
    }
    
    wg.Wait()
}

func (eb *EventBus) PublishAsync(event Event) {
    go eb.Publish(event)
}

// EventStream provides filtered event streaming
type EventStream struct {
    bus       *EventBus
    filters   []EventFilter
    transform EventTransformer
}

type EventFilter func(Event) bool
type EventTransformer func(Event) Event

func (es *EventStream) Subscribe(subscriberID string, eventTypes ...EventType) <-chan Event {
    merged := make(chan Event, es.bus.bufferSize)
    
    // Subscribe to each event type
    for _, eventType := range eventTypes {
        ch := es.bus.Subscribe(subscriberID, eventType, es.bus.bufferSize)
        
        go func(source <-chan Event) {
            for event := range source {
                // Apply filters
                if es.passesFilters(event) {
                    // Apply transformation
                    if es.transform != nil {
                        event = es.transform(event)
                    }
                    
                    select {
                    case merged <- event:
                    default:
                        // Drop if merged channel is full
                    }
                }
            }
        }(ch)
    }
    
    return merged
}

func (es *EventStream) passesFilters(event Event) bool {
    for _, filter := range es.filters {
        if !filter(event) {
            return false
        }
    }
    return true
}
```

### Server-Sent Events (SSE) Pattern

```go
// sse_server.go
package statusline

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type SSEServer struct {
    clients   map[string]*SSEClient
    mu        sync.RWMutex
    eventBus  *EventBus
    heartbeat time.Duration
}

type SSEClient struct {
    id       string
    events   chan []byte
    close    chan struct{}
    closed   bool
}

func NewSSEServer(eventBus *EventBus) *SSEServer {
    return &SSEServer{
        clients:   make(map[string]*SSEClient),
        eventBus:  eventBus,
        heartbeat: 30 * time.Second,
    }
}

func (s *SSEServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Check if client accepts SSE
    if r.Header.Get("Accept") != "text/event-stream" {
        http.Error(w, "SSE not supported", http.StatusBadRequest)
        return
    }
    
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering
    
    // Create client
    client := &SSEClient{
        id:     generateClientID(),
        events: make(chan []byte, 100),
        close:  make(chan struct{}),
    }
    
    s.addClient(client)
    defer s.removeClient(client.id)
    
    // Subscribe to events
    eventChan := s.eventBus.Subscribe(client.id, EventStatusUpdate, 10)
    defer s.eventBus.Unsubscribe(client.id, EventStatusUpdate)
    
    // Start event forwarding
    go s.forwardEvents(client, eventChan)
    
    // Flush immediately to establish connection
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    
    // Send initial connection event
    s.sendEvent(w, flusher, "connected", map[string]string{"clientId": client.id})
    
    // Main event loop
    ticker := time.NewTicker(s.heartbeat)
    defer ticker.Stop()
    
    for {
        select {
        case event := <-client.events:
            // Send event
            if _, err := w.Write(event); err != nil {
                return
            }
            flusher.Flush()
            
        case <-ticker.C:
            // Send heartbeat
            s.sendEvent(w, flusher, "ping", map[string]int64{"timestamp": time.Now().Unix()})
            
        case <-r.Context().Done():
            // Client disconnected
            return
            
        case <-client.close:
            // Server closing connection
            return
        }
    }
}

func (s *SSEServer) forwardEvents(client *SSEClient, events <-chan Event) {
    for event := range events {
        data, err := json.Marshal(event.Data)
        if err != nil {
            continue
        }
        
        sseEvent := fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, data)
        
        select {
        case client.events <- []byte(sseEvent):
        case <-client.close:
            return
        default:
            // Buffer full, skip event
        }
    }
}

func (s *SSEServer) sendEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
    jsonData, _ := json.Marshal(data)
    fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
    flusher.Flush()
}

func (s *SSEServer) Broadcast(event Event) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    data, err := json.Marshal(event)
    if err != nil {
        return
    }
    
    message := fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, data)
    
    for _, client := range s.clients {
        select {
        case client.events <- []byte(message):
        default:
            // Client buffer full
        }
    }
}

func (s *SSEServer) addClient(client *SSEClient) {
    s.mu.Lock()
    s.clients[client.id] = client
    s.mu.Unlock()
}

func (s *SSEServer) removeClient(id string) {
    s.mu.Lock()
    if client, ok := s.clients[id]; ok {
        if !client.closed {
            close(client.close)
            client.closed = true
        }
        delete(s.clients, id)
    }
    s.mu.Unlock()
}
```

### WebSocket Pattern

```go
// websocket_server.go
package statusline

import (
    "context"
    "encoding/json"
    "net/http"
    "sync"
    "time"
    
    "github.com/gorilla/websocket"
)

type WSMessage struct {
    Type    string          `json:"type"`
    Data    json.RawMessage `json:"data"`
    ID      string          `json:"id,omitempty"`
}

type WSHub struct {
    clients    map[*WSClient]bool
    broadcast  chan []byte
    register   chan *WSClient
    unregister chan *WSClient
    mu         sync.RWMutex
}

type WSClient struct {
    hub       *WSHub
    conn      *websocket.Conn
    send      chan []byte
    id        string
    filters   []string
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        // Configure origin checking in production
        return true
    },
}

func NewWSHub() *WSHub {
    return &WSHub{
        clients:    make(map[*WSClient]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *WSClient),
        unregister: make(chan *WSClient),
    }
}

func (h *WSHub) Run() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            
            // Send welcome message
            welcome := WSMessage{
                Type: "connected",
                Data: json.RawMessage(`{"status":"connected"}`),
            }
            if msg, err := json.Marshal(welcome); err == nil {
                client.send <- msg
            }
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            
        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    // Client's buffer is full, close it
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
            
        case <-ticker.C:
            // Send ping to all clients
            ping := WSMessage{Type: "ping"}
            if msg, err := json.Marshal(ping); err == nil {
                h.mu.RLock()
                for client := range h.clients {
                    select {
                    case client.send <- msg:
                    default:
                    }
                }
                h.mu.RUnlock()
            }
        }
    }
}

func (h *WSHub) ServeWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    client := &WSClient{
        hub:  h,
        conn: conn,
        send: make(chan []byte, 256),
        id:   generateClientID(),
    }
    
    h.register <- client
    
    // Start goroutines
    go client.writePump()
    go client.readPump()
}

func (c *WSClient) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            c.conn.WriteMessage(websocket.TextMessage, message)
            
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (c *WSClient) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    c.conn.SetReadLimit(512)
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        
        // Process message
        var msg WSMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }
        
        c.handleMessage(msg)
    }
}

func (c *WSClient) handleMessage(msg WSMessage) {
    switch msg.Type {
    case "subscribe":
        var filters []string
        json.Unmarshal(msg.Data, &filters)
        c.filters = filters
        
    case "command":
        // Handle commands
        var cmd Command
        json.Unmarshal(msg.Data, &cmd)
        c.processCommand(cmd)
    }
}
```

## Performance Optimization

### Memory Pooling

```go
// memory_pool.go
package statusline

import (
    "bytes"
    "sync"
)

// BufferPool reduces allocations for rendering
var bufferPool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}

func getBuffer() *bytes.Buffer {
    buf := bufferPool.Get().(*bytes.Buffer)
    buf.Reset()
    return buf
}

func putBuffer(buf *bytes.Buffer) {
    if buf.Cap() > 64*1024 { // Don't pool large buffers
        return
    }
    bufferPool.Put(buf)
}

// WidgetPool for reusing widget instances
type WidgetPool struct {
    pools map[string]*sync.Pool
    mu    sync.RWMutex
}

func NewWidgetPool() *WidgetPool {
    return &WidgetPool{
        pools: make(map[string]*sync.Pool),
    }
}

func (wp *WidgetPool) Get(widgetType string) Widget {
    wp.mu.RLock()
    pool, ok := wp.pools[widgetType]
    wp.mu.RUnlock()
    
    if !ok {
        return nil
    }
    
    if widget := pool.Get(); widget != nil {
        return widget.(Widget)
    }
    
    return nil
}

func (wp *WidgetPool) Put(widget Widget) {
    widgetType := widget.Type()
    
    wp.mu.Lock()
    if _, ok := wp.pools[widgetType]; !ok {
        wp.pools[widgetType] = &sync.Pool{
            New: func() interface{} {
                return widget.Clone()
            },
        }
    }
    wp.mu.Unlock()
    
    widget.Reset()
    wp.pools[widgetType].Put(widget)
}

// Usage in rendering
func (s *StatusLine) renderOptimized() string {
    buf := getBuffer()
    defer putBuffer(buf)
    
    for i, widget := range s.widgets {
        if i > 0 {
            buf.WriteString(" | ")
        }
        
        // Try to get from pool
        pooled := s.widgetPool.Get(widget.Type())
        if pooled != nil {
            pooled.Configure(widget.Config())
            widget = pooled
            defer s.widgetPool.Put(widget)
        }
        
        content, _ := widget.Render(s.ctx)
        buf.WriteString(content)
    }
    
    return buf.String()
}
```

### Zero-Allocation Techniques

```go
// zero_alloc.go
package statusline

import (
    "sync"
    "unsafe"
)

// StringIntern reduces string allocations
type StringIntern struct {
    mu    sync.RWMutex
    table map[string]string
}

func NewStringIntern() *StringIntern {
    return &StringIntern{
        table: make(map[string]string),
    }
}

func (si *StringIntern) Intern(s string) string {
    si.mu.RLock()
    if interned, ok := si.table[s]; ok {
        si.mu.RUnlock()
        return interned
    }
    si.mu.RUnlock()
    
    si.mu.Lock()
    defer si.mu.Unlock()
    
    // Double-check
    if interned, ok := si.table[s]; ok {
        return interned
    }
    
    si.table[s] = s
    return s
}

// RingBuffer for zero-allocation circular buffer
type RingBuffer struct {
    data []byte
    size int
    head int
    tail int
    mu   sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
    return &RingBuffer{
        data: make([]byte, size),
        size: size,
    }
}

func (rb *RingBuffer) Write(p []byte) (n int, err error) {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    
    n = len(p)
    if n > rb.size {
        p = p[n-rb.size:]
        n = rb.size
    }
    
    for i := 0; i < n; i++ {
        rb.data[rb.head] = p[i]
        rb.head = (rb.head + 1) % rb.size
        if rb.head == rb.tail {
            rb.tail = (rb.tail + 1) % rb.size
        }
    }
    
    return n, nil
}

func (rb *RingBuffer) Read(p []byte) (n int, err error) {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    
    if rb.head == rb.tail {
        return 0, nil
    }
    
    if rb.tail < rb.head {
        n = copy(p, rb.data[rb.tail:rb.head])
    } else {
        n = copy(p, rb.data[rb.tail:])
        if n < len(p) {
            n += copy(p[n:], rb.data[:rb.head])
        }
    }
    
    rb.tail = (rb.tail + n) % rb.size
    return n, nil
}

// Pre-allocated slice reuse
type WidgetCache struct {
    widgets []Widget
    results []string
    mu      sync.Mutex
}

func NewWidgetCache(capacity int) *WidgetCache {
    return &WidgetCache{
        widgets: make([]Widget, 0, capacity),
        results: make([]string, 0, capacity),
    }
}

func (wc *WidgetCache) RenderAll(ctx context.Context) []string {
    wc.mu.Lock()
    defer wc.mu.Unlock()
    
    // Reuse existing slice, just reset length
    wc.results = wc.results[:0]
    
    for _, widget := range wc.widgets {
        result, _ := widget.Render(ctx)
        wc.results = append(wc.results, result)
    }
    
    return wc.results
}
```

### Lock-Free Atomic Operations

```go
// atomic_ops.go
package statusline

import (
    "sync/atomic"
    "unsafe"
)

// AtomicStatus for lock-free status updates
type AtomicStatus struct {
    value unsafe.Pointer // *StatusData
}

type StatusData struct {
    SessionID string
    Active    bool
    Tokens    int64
    Updated   int64 // Unix timestamp
}

func (as *AtomicStatus) Load() *StatusData {
    return (*StatusData)(atomic.LoadPointer(&as.value))
}

func (as *AtomicStatus) Store(status *StatusData) {
    atomic.StorePointer(&as.value, unsafe.Pointer(status))
}

func (as *AtomicStatus) CompareAndSwap(old, new *StatusData) bool {
    return atomic.CompareAndSwapPointer(
        &as.value,
        unsafe.Pointer(old),
        unsafe.Pointer(new),
    )
}

// Atomic counter collection
type AtomicCounters struct {
    renders      uint64
    errors       uint64
    cacheHits    uint64
    cacheMisses  uint64
}

func (ac *AtomicCounters) IncrementRenders() uint64 {
    return atomic.AddUint64(&ac.renders, 1)
}

func (ac *AtomicCounters) IncrementErrors() uint64 {
    return atomic.AddUint64(&ac.errors, 1)
}

func (ac *AtomicCounters) RecordCacheAccess(hit bool) {
    if hit {
        atomic.AddUint64(&ac.cacheHits, 1)
    } else {
        atomic.AddUint64(&ac.cacheMisses, 1)
    }
}

func (ac *AtomicCounters) GetStats() (renders, errors, hits, misses uint64) {
    return atomic.LoadUint64(&ac.renders),
           atomic.LoadUint64(&ac.errors),
           atomic.LoadUint64(&ac.cacheHits),
           atomic.LoadUint64(&ac.cacheMisses)
}

// Lock-free queue implementation
type LockFreeQueue struct {
    head unsafe.Pointer // *node
    tail unsafe.Pointer // *node
}

type node struct {
    value interface{}
    next  unsafe.Pointer // *node
}

func NewLockFreeQueue() *LockFreeQueue {
    n := &node{}
    return &LockFreeQueue{
        head: unsafe.Pointer(n),
        tail: unsafe.Pointer(n),
    }
}

func (q *LockFreeQueue) Enqueue(v interface{}) {
    n := &node{value: v}
    for {
        tail := (*node)(atomic.LoadPointer(&q.tail))
        next := (*node)(atomic.LoadPointer(&tail.next))
        
        if tail == (*node)(atomic.LoadPointer(&q.tail)) {
            if next == nil {
                if atomic.CompareAndSwapPointer(&tail.next, nil, unsafe.Pointer(n)) {
                    atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(n))
                    break
                }
            } else {
                atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
            }
        }
    }
}

func (q *LockFreeQueue) Dequeue() (interface{}, bool) {
    for {
        head := (*node)(atomic.LoadPointer(&q.head))
        tail := (*node)(atomic.LoadPointer(&q.tail))
        next := (*node)(atomic.LoadPointer(&head.next))
        
        if head == (*node)(atomic.LoadPointer(&q.head)) {
            if head == tail {
                if next == nil {
                    return nil, false
                }
                atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
            } else {
                v := next.value
                if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
                    return v, true
                }
            }
        }
    }
}
```

### Batch Processing

```go
// batch_processing.go
package statusline

import (
    "context"
    "sync"
    "time"
)

// BatchRenderer optimizes rendering by batching updates
type BatchRenderer struct {
    batchSize    int
    flushInterval time.Duration
    items        []RenderItem
    mu           sync.Mutex
    flushCh      chan struct{}
    processor    func([]RenderItem) error
}

type RenderItem struct {
    Widget   Widget
    Priority int
    Result   chan<- string
}

func NewBatchRenderer(batchSize int, flushInterval time.Duration) *BatchRenderer {
    br := &BatchRenderer{
        batchSize:     batchSize,
        flushInterval: flushInterval,
        items:         make([]RenderItem, 0, batchSize),
        flushCh:       make(chan struct{}, 1),
    }
    
    go br.flushLoop()
    return br
}

func (br *BatchRenderer) Add(item RenderItem) error {
    br.mu.Lock()
    br.items = append(br.items, item)
    shouldFlush := len(br.items) >= br.batchSize
    br.mu.Unlock()
    
    if shouldFlush {
        br.triggerFlush()
    }
    
    return nil
}

func (br *BatchRenderer) triggerFlush() {
    select {
    case br.flushCh <- struct{}{}:
    default:
    }
}

func (br *BatchRenderer) flushLoop() {
    ticker := time.NewTicker(br.flushInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            br.flush()
        case <-br.flushCh:
            br.flush()
        }
    }
}

func (br *BatchRenderer) flush() {
    br.mu.Lock()
    if len(br.items) == 0 {
        br.mu.Unlock()
        return
    }
    
    batch := br.items
    br.items = make([]RenderItem, 0, br.batchSize)
    br.mu.Unlock()
    
    // Process batch in parallel
    var wg sync.WaitGroup
    for _, item := range batch {
        wg.Add(1)
        go func(ri RenderItem) {
            defer wg.Done()
            
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
            defer cancel()
            
            result, _ := ri.Widget.Render(ctx)
            ri.Result <- result
        }(item)
    }
    
    wg.Wait()
}

// CoalescingProcessor reduces redundant updates
type CoalescingProcessor struct {
    window    time.Duration
    processor func([]Update) error
    pending   map[string]Update
    mu        sync.Mutex
    timer     *time.Timer
}

type Update struct {
    ID        string
    Data      interface{}
    Timestamp time.Time
}

func NewCoalescingProcessor(window time.Duration, processor func([]Update) error) *CoalescingProcessor {
    return &CoalescingProcessor{
        window:    window,
        processor: processor,
        pending:   make(map[string]Update),
    }
}

func (cp *CoalescingProcessor) Add(update Update) {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    // Coalesce updates with same ID
    cp.pending[update.ID] = update
    
    // Start or reset timer
    if cp.timer != nil {
        cp.timer.Stop()
    }
    
    cp.timer = time.AfterFunc(cp.window, cp.process)
}

func (cp *CoalescingProcessor) process() {
    cp.mu.Lock()
    
    if len(cp.pending) == 0 {
        cp.mu.Unlock()
        return
    }
    
    // Extract updates
    updates := make([]Update, 0, len(cp.pending))
    for _, update := range cp.pending {
        updates = append(updates, update)
    }
    cp.pending = make(map[string]Update)
    
    cp.mu.Unlock()
    
    // Process coalesced updates
    if err := cp.processor(updates); err != nil {
        // Handle error
    }
}
```

### Performance Monitoring

```go
// performance_monitor.go
package statusline

import (
    "context"
    "runtime"
    "sync"
    "time"
)

// PerformanceMonitor tracks statusline performance metrics
type PerformanceMonitor struct {
    renderTimes   *TimingHistogram
    widgetTimes   map[string]*TimingHistogram
    goroutines    []int
    memStats      []runtime.MemStats
    mu            sync.RWMutex
}

type TimingHistogram struct {
    buckets  []time.Duration
    counts   []uint64
    total    uint64
    sum      time.Duration
}

func NewPerformanceMonitor() *PerformanceMonitor {
    return &PerformanceMonitor{
        renderTimes: NewTimingHistogram([]time.Duration{
            1 * time.Millisecond,
            5 * time.Millisecond,
            10 * time.Millisecond,
            25 * time.Millisecond,
            50 * time.Millisecond,
            100 * time.Millisecond,
        }),
        widgetTimes: make(map[string]*TimingHistogram),
    }
}

func (pm *PerformanceMonitor) RecordRender(duration time.Duration) {
    pm.renderTimes.Record(duration)
}

func (pm *PerformanceMonitor) RecordWidget(widgetID string, duration time.Duration) {
    pm.mu.Lock()
    if _, ok := pm.widgetTimes[widgetID]; !ok {
        pm.widgetTimes[widgetID] = NewTimingHistogram([]time.Duration{
            100 * time.Microsecond,
            500 * time.Microsecond,
            1 * time.Millisecond,
            5 * time.Millisecond,
            10 * time.Millisecond,
        })
    }
    pm.mu.Unlock()
    
    pm.widgetTimes[widgetID].Record(duration)
}

func (pm *PerformanceMonitor) StartMonitoring(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            pm.collect()
        case <-ctx.Done():
            return
        }
    }
}

func (pm *PerformanceMonitor) collect() {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    // Collect goroutine count
    pm.goroutines = append(pm.goroutines, runtime.NumGoroutine())
    if len(pm.goroutines) > 360 { // Keep 1 hour of data
        pm.goroutines = pm.goroutines[1:]
    }
    
    // Collect memory stats
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    pm.memStats = append(pm.memStats, m)
    if len(pm.memStats) > 360 {
        pm.memStats = pm.memStats[1:]
    }
}

func (pm *PerformanceMonitor) GetReport() PerformanceReport {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    report := PerformanceReport{
        RenderP50:  pm.renderTimes.Percentile(0.5),
        RenderP95:  pm.renderTimes.Percentile(0.95),
        RenderP99:  pm.renderTimes.Percentile(0.99),
    }
    
    // Widget timings
    report.WidgetTimings = make(map[string]WidgetTiming)
    for id, hist := range pm.widgetTimes {
        report.WidgetTimings[id] = WidgetTiming{
            P50: hist.Percentile(0.5),
            P95: hist.Percentile(0.95),
            P99: hist.Percentile(0.99),
        }
    }
    
    // Resource usage
    if len(pm.goroutines) > 0 {
        report.AvgGoroutines = average(pm.goroutines)
        report.MaxGoroutines = max(pm.goroutines)
    }
    
    if len(pm.memStats) > 0 {
        report.AvgHeapAlloc = averageUint64(extractHeapAlloc(pm.memStats))
        report.MaxHeapAlloc = maxUint64(extractHeapAlloc(pm.memStats))
    }
    
    return report
}

type PerformanceReport struct {
    RenderP50     time.Duration
    RenderP95     time.Duration
    RenderP99     time.Duration
    WidgetTimings map[string]WidgetTiming
    AvgGoroutines int
    MaxGoroutines int
    AvgHeapAlloc  uint64
    MaxHeapAlloc  uint64
}

type WidgetTiming struct {
    P50 time.Duration
    P95 time.Duration
    P99 time.Duration
}

func NewTimingHistogram(buckets []time.Duration) *TimingHistogram {
    return &TimingHistogram{
        buckets: buckets,
        counts:  make([]uint64, len(buckets)+1),
    }
}

func (th *TimingHistogram) Record(d time.Duration) {
    atomic.AddUint64(&th.total, 1)
    atomic.AddInt64((*int64)(&th.sum), int64(d))
    
    // Find bucket
    for i, bucket := range th.buckets {
        if d <= bucket {
            atomic.AddUint64(&th.counts[i], 1)
            return
        }
    }
    
    // Overflow bucket
    atomic.AddUint64(&th.counts[len(th.counts)-1], 1)
}

func (th *TimingHistogram) Percentile(p float64) time.Duration {
    total := atomic.LoadUint64(&th.total)
    if total == 0 {
        return 0
    }
    
    target := uint64(float64(total) * p)
    cumulative := uint64(0)
    
    for i, count := range th.counts {
        cumulative += atomic.LoadUint64(&count)
        if cumulative >= target {
            if i < len(th.buckets) {
                return th.buckets[i]
            }
            // In overflow bucket, return average
            return time.Duration(atomic.LoadInt64((*int64)(&th.sum)) / int64(total))
        }
    }
    
    return 0
}
```

## Best Practices Summary

### For <10ms Render Latency

1. **Parallel Widget Rendering**
   - Use worker pools with pre-allocated goroutines
   - Limit workers to runtime.NumCPU()
   - Apply per-widget timeouts (5ms max)

2. **Bounded Concurrency**
   - Use semaphores to limit concurrent operations
   - Implement backpressure on all channels
   - Set buffer sizes based on expected load

3. **Memory Optimization**
   - Use sync.Pool for buffers and temporary objects
   - Pre-allocate slices with appropriate capacity
   - Intern frequently used strings

4. **Lock-Free Where Possible**
   - Use atomic operations for counters and flags
   - Implement lock-free queues for high-throughput paths
   - Use sync.Map for read-heavy concurrent maps

5. **Context Best Practices**
   - Always propagate context through call chains
   - Set aggressive timeouts (10ms for full render)
   - Use context values for request tracing

6. **Avoid Allocations**
   - Reuse buffers and slices
   - Use zero-copy techniques where possible
   - Batch small operations to reduce overhead

7. **Monitor Everything**
   - Track goroutine counts
   - Measure render times per widget
   - Monitor memory allocations
   - Use pprof for production profiling

### Common Pitfalls to Avoid

1. **Unbounded Goroutine Creation**
   - Always use worker pools or semaphores
   - Set limits on concurrent operations
   - Monitor goroutine counts in production

2. **Missing Context Cancellation**
   - Every goroutine must respect context
   - Use context for all I/O operations
   - Implement proper cleanup on cancellation

3. **Channel Leaks**
   - Only producers should close channels
   - Always drain channels before abandoning
   - Use buffered channels carefully

4. **Race Conditions**
   - Run tests with -race flag
   - Use proper synchronization primitives
   - Avoid sharing memory without protection

5. **Blocking Operations**
   - Always provide timeouts or cancellation
   - Use non-blocking channel operations where appropriate
   - Implement circuit breakers for external calls

6. **Memory Leaks**
   - Clean up all resources with defer
   - Use cleanup stacks for complex initialization
   - Test for leaks with goleak

7. **Poor Error Handling**
   - Propagate errors through channels
   - Use errgroup for coordinated error handling
   - Log errors with appropriate context

## Example: High-Performance Statusline

```go
// statusline.go
package statusline

import (
    "context"
    "fmt"
    "io"
    "runtime"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

// StatusLine represents the main statusline manager
type StatusLine struct {
    // Core components
    widgets      []Widget
    renderer     *ParallelRenderer
    output       io.Writer
    
    // Concurrency control
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
    
    // Configuration
    updateInterval time.Duration
    renderTimeout  time.Duration
    
    // Performance tracking
    metrics      *Metrics
    perfMonitor  *PerformanceMonitor
}

// Metrics tracks statusline performance
type Metrics struct {
    renders      uint64
    errors       uint64
    totalTime    int64 // nanoseconds
    maxRender    int64 // nanoseconds
}

// NewStatusLine creates an optimized statusline
func NewStatusLine(widgets []Widget, output io.Writer) *StatusLine {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &StatusLine{
        widgets:        widgets,
        renderer:       NewParallelRenderer(runtime.NumCPU()),
        output:         output,
        ctx:            ctx,
        cancel:         cancel,
        updateInterval: 300 * time.Millisecond,
        renderTimeout:  10 * time.Millisecond,
        metrics:        &Metrics{},
        perfMonitor:    NewPerformanceMonitor(),
    }
}

// Start begins the statusline update loop
func (s *StatusLine) Start() {
    s.wg.Add(2)
    go s.updateLoop()
    go s.perfMonitor.StartMonitoring(s.ctx)
}

// Stop gracefully shuts down the statusline
func (s *StatusLine) Stop() error {
    s.cancel()
    
    // Wait with timeout
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(5 * time.Second):
        return ErrShutdownTimeout
    }
}

// updateLoop runs the main render cycle
func (s *StatusLine) updateLoop() {
    defer s.wg.Done()
    
    ticker := time.NewTicker(s.updateInterval)
    defer ticker.Stop()
    
    // Render immediately on start
    s.render()
    
    for {
        select {
        case <-ticker.C:
            s.render()
        case <-s.ctx.Done():
            return
        }
    }
}

// render performs a single render cycle
func (s *StatusLine) render() {
    start := time.Now()
    
    // Create render context with timeout
    ctx, cancel := context.WithTimeout(s.ctx, s.renderTimeout)
    defer cancel()
    
    // Add request ID for tracing
    ctx = WithRequestID(ctx, generateRequestID())
    
    // Render all widgets in parallel
    results, err := s.renderer.RenderAll(ctx, s.widgets)
    
    duration := time.Since(start)
    
    // Update metrics
    atomic.AddUint64(&s.metrics.renders, 1)
    atomic.AddInt64(&s.metrics.totalTime, int64(duration))
    
    // Update max render time
    for {
        old := atomic.LoadInt64(&s.metrics.maxRender)
        if duration.Nanoseconds() <= old {
            break
        }
        if atomic.CompareAndSwapInt64(&s.metrics.maxRender, old, duration.Nanoseconds()) {
            break
        }
    }
    
    // Record in performance monitor
    s.perfMonitor.RecordRender(duration)
    
    if err != nil {
        atomic.AddUint64(&s.metrics.errors, 1)
        s.displayError(err)
        return
    }
    
    // Combine and display results
    output := s.combineResults(results)
    s.display(output)
}

// combineResults merges widget outputs
func (s *StatusLine) combineResults(results []RenderResult) string {
    buf := getBuffer()
    defer putBuffer(buf)
    
    first := true
    for _, result := range results {
        if result.Error != nil {
            continue
        }
        
        if !first {
            buf.WriteString(" | ")
        }
        first = false
        
        buf.WriteString(result.Content)
    }
    
    return buf.String()
}

// display writes output to terminal
func (s *StatusLine) display(content string) {
    // Clear line and write new content
    fmt.Fprintf(s.output, "\r\033[K%s", content)
}

// displayError shows error state
func (s *StatusLine) displayError(err error) {
    fmt.Fprintf(s.output, "\r\033[K[ERROR: %v]", err)
}

// GetMetrics returns current performance metrics
func (s *StatusLine) GetMetrics() StatusLineMetrics {
    renders := atomic.LoadUint64(&s.metrics.renders)
    errors := atomic.LoadUint64(&s.metrics.errors)
    totalTime := time.Duration(atomic.LoadInt64(&s.metrics.totalTime))
    maxRender := time.Duration(atomic.LoadInt64(&s.metrics.maxRender))
    
    avgRender := time.Duration(0)
    if renders > 0 {
        avgRender = totalTime / time.Duration(renders)
    }
    
    return StatusLineMetrics{
        TotalRenders: renders,
        TotalErrors:  errors,
        AvgRender:    avgRender,
        MaxRender:    maxRender,
        Performance:  s.perfMonitor.GetReport(),
    }
}

// ParallelRenderer handles concurrent widget rendering
type ParallelRenderer struct {
    workers    int
    pool       *WorkerPool
    bufferPool *sync.Pool
}

func NewParallelRenderer(workers int) *ParallelRenderer {
    return &ParallelRenderer{
        workers: workers,
        pool:    NewWorkerPool(workers),
        bufferPool: &sync.Pool{
            New: func() interface{} {
                return new(bytes.Buffer)
            },
        },
    }
}

// RenderAll renders all widgets with <10ms target
func (r *ParallelRenderer) RenderAll(ctx context.Context, widgets []Widget) ([]RenderResult, error) {
    results := make([]RenderResult, len(widgets))
    resultChan := make(chan IndexedResult, len(widgets))
    
    // Use semaphore for concurrency control
    sem := make(chan struct{}, r.workers)
    
    var wg sync.WaitGroup
    for i, widget := range widgets {
        wg.Add(1)
        
        go func(idx int, w Widget) {
            defer wg.Done()
            
            // Acquire semaphore
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                resultChan <- IndexedResult{
                    Index: idx,
                    Result: RenderResult{
                        WidgetID: w.ID(),
                        Error:    ctx.Err(),
                    },
                }
                return
            }
            
            // Create widget context with 5ms timeout
            wctx, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
            defer cancel()
            
            // Track widget render time
            start := time.Now()
            
            // Render with panic recovery
            var content string
            var err error
            
            func() {
                defer func() {
                    if r := recover(); r != nil {
                        err = fmt.Errorf("panic in widget %s: %v", w.ID(), r)
                    }
                }()
                
                content, err = w.Render(wctx)
            }()
            
            duration := time.Since(start)
            
            resultChan <- IndexedResult{
                Index: idx,
                Result: RenderResult{
                    WidgetID: w.ID(),
                    Content:  content,
                    Error:    err,
                    Duration: duration,
                },
            }
        }(i, widget)
    }
    
    // Wait for all renders to complete
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Collect results in order
    for res := range resultChan {
        results[res.Index] = res.Result
    }
    
    // Check if we exceeded timeout
    select {
    case <-ctx.Done():
        return nil, fmt.Errorf("render timeout: %w", ctx.Err())
    default:
    }
    
    return results, nil
}

type IndexedResult struct {
    Index  int
    Result RenderResult
}

type RenderResult struct {
    WidgetID string
    Content  string
    Error    error
    Duration time.Duration
}

// Widget interface for statusline components
type Widget interface {
    ID() string
    Type() string
    Render(context.Context) (string, error)
    Configure(interface{}) error
    Clone() Widget
    Reset()
    Config() interface{}
}

// Usage example
func main() {
    // Create widgets
    widgets := []Widget{
        NewClaudeWidget(),
        NewTokenWidget(),
        NewTimeWidget(),
        NewGitWidget(),
    }
    
    // Create statusline
    sl := NewStatusLine(widgets, os.Stdout)
    
    // Handle shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    // Start statusline
    sl.Start()
    
    // Wait for shutdown
    <-sigChan
    
    // Graceful shutdown
    if err := sl.Stop(); err != nil {
        log.Printf("Shutdown error: %v", err)
        os.Exit(1)
    }
    
    // Print final metrics
    metrics := sl.GetMetrics()
    log.Printf("Final metrics: %+v", metrics)
}
```

This comprehensive documentation provides all the Go concurrency patterns needed for implementing a high-performance statusline with <10ms render latency. The patterns focus on efficiency, proper resource management, preventing common pitfalls, and maintaining clean, idiomatic Go code while achieving real-time performance requirements.