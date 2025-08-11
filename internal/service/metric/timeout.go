package metric

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/slok/grafterm/internal/model"
)

// TimeoutConfig defines timeout configuration for metric queries
const (
	DefaultTimeout      = 5 * time.Second
	MaxConcurrentCalls  = 10
	MaxRetransmission   = 3
)

// QueryExecutor handles metric queries with proper timeout management
type QueryExecutor struct {
	semaphore chan struct{}
	cache     *MetricCache
	metrics   *ExecutionMetrics
}

// NewQueryExecutor creates a new query executor with timeout management
func NewQueryExecutor(cache *MetricCache) *QueryExecutor {
	return &QueryExecutor{
		semaphore: make(chan struct{}, MaxConcurrentCalls),
		cache:     cache,
		metrics:   NewExecutionMetrics(),
	}
}

// ExecuteQuery performs a metric query with context timeout
func (qe *QueryExecutor) ExecuteQuery(
	ctx context.Context,
	gatherer Gatherer,
	query string,
	tr model.TimeRange,
) ([]model.MetricSeries, error) {
	// Check cache first
	cacheKey := NewCacheKey(gatherer.ID(), query, tr)
	if cached, found := qe.cache.Get(cacheKey); found {
		qe.metrics.RecordCacheHit()
		return cached, nil
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	// Rate limiting with semaphore
	select {
	case qe.semaphore <- struct{}{}:
		defer func() { <-qe.semaphore }()
	case <-ctx.Done():
		return nil, fmt.Errorf("query execution timeout waiting for rate limit: %w", ctx.Err())
	}

	// Execute query with retry logic
	result, err := qe.executeWithRetry(ctx, gatherer, query, tr)
	if err != nil {
		qe.metrics.RecordError(err)
		return nil, err
	}

	// Cache successful results
	qe.cache.Set(cacheKey, result)
	qe.metrics.RecordSuccess()

	return result, nil
}

// executeWithRetry implements retry logic for failed queries
func (qe *QueryExecutor) executeWithRetry(
	ctx context.Context,
	gatherer Gatherer,
	query string,
	tr model.TimeRange,
) ([]model.MetricSeries, error) {
	var lastErr error
	
	for attempt := 0; attempt < MaxRetransmission; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff for retries
		if attempt > 0 {
			sleepTime := time.Duration(attempt*attempt*100) * time.Millisecond
			select {
			case <-time.After(sleepTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		result, err := gatherer.GatherSingle(ctx, query, tr)
		if err == nil {
			return result, nil
		}

		lastErr = err
		
		// Handle specific error types differently
		if isContextError(err) {
			return nil, err // Don't retry context errors
		}

		// Check if context expired
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("query failed after %d attempts: %w", MaxRetransmission, lastErr)
}

// isContextError checks if error is related to context cancellation/timeout
func isContextError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == context.DeadlineExceeded.Error() || 
	       err.Error() == context.Canceled.Error()
}

// ParallelQueryExecutor for concurrent widget execution
type ParallelQueryExecutor struct {
	qe *QueryExecutor
}

// NewParallelQueryExecutor creates parallel executor
func NewParallelQueryExecutor(qe *QueryExecutor) *ParallelQueryExecutor {
	return &ParallelQueryExecutor{qe: qe}
}

// ExecuteWidgetQueries processes multiple widgets concurrently
func (pqe *ParallelQueryExecutor) ExecuteWidgetQueries(
	ctx context.Context,
	widgets []WidgetData,
) map[string]WidgetResult {
	var wg sync.WaitGroup
	results := make(map[string]WidgetResult, len(widgets))
	resultsCh := make(chan WidgetResult, len(widgets))

	for _, widget := range widgets {
		wg.Add(1)
		go func(w WidgetData) {
			defer wg.Done()
			
			widgetCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
			defer cancel()

			metrics, err := pqe.qe.ExecuteQuery(widgetCtx, w.Gatherer, w.Query, w.TimeRange)
			
			result := WidgetResult{
				ID:      w.ID,
				Metrics: metrics,
				Error:   err,
			}
			
			// Non-blocking send for graceful shutdown
			select {
			case resultsCh <- result:
			case <-ctx.Done():
				return
			}
		}(widget)
	}

	// Close channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	for result := range resultsCh {
		results[result.ID] = result
	}

	return results
}

// WidgetData represents a single widget query
type WidgetData struct {
	ID       string
	Gatherer Gatherer
	Query    string
	TimeRange model.TimeRange
}

// WidgetResult contains the execution result for a widget
type WidgetResult struct {
	ID      string
	Metrics []model.MetricSeries
	Error   error
}