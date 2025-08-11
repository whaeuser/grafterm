package prometheus

import (
	"context"
	"fmt"
	"sync"
	"time"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"github.com/slok/grafterm/internal/model"
)

// EnhancedGatherer provides improved Prometheus integration with timeout management
type EnhancedGatherer interface {
	metric.Gatherer
	ID() string
	SetTimeout(duration time.Duration)
	GetLastExecutionTime() time.Duration
}

// enhancedGatherer wraps the standard prometheus gatherer with enhanced features
type enhancedGatherer struct {
	base          *gatherer
	id            string
	mu            sync.RWMutex
	timeout       time.Duration
	lastExecTime  time.Duration
	metrics       *gathererMetrics
}

// gathererMetrics tracks execution statistics
type gathererMetrics struct {
	queriesTotal      int64
	queriesSuccessful int64
	queriesFailed     int64
	queriesTimeout   int64
	averageExecTime   time.Duration
	mu                sync.RWMutex
}

// NewEnhancedGatherer returns an enhanced version of the Prometheus gatherer
func NewEnhancedGatherer(cfg ConfigGatherer, datasourceID string) EnhancedGatherer {
	return &enhancedGatherer{
		base:    &gatherer{cli: cfg.Client, cfg: cfg},
		id:      datasourceID,
		timeout: DefaultTimeout,
		metrics: &gathererMetrics{},
	}
}

func (eg *enhancedGatherer) ID() string {
	return eg.id
}

func (eg *enhancedGatherer) SetTimeout(duration time.Duration) {
	if duration <= 0 {
		duration = DefaultTimeout
	}
	
	// Cap timeout at 30s to prevent excessive waits
	if duration > 30*time.Second {
		duration = 30 * time.Second
	}
	
	// Enforce minimum timeout of 1s
	if duration < time.Second {
		duration = time.Second
	}
	
	eg.mu.Lock()
	defer eg.mu.Unlock()
	eg.timeout = duration
}

func (eg *enhancedGatherer) GetLastExecutionTime() time.Duration {
	return eg.getLastExecutionTime()
}

// GatherSingle gathers a single metric point with timeout management
func (eg *enhancedGatherer) GatherSingle(ctx context.Context, query string, t time.Time) ([]model.MetricSeries, error) {
	start := time.Now()
	defer func() {
		if eg != nil {
			eg.recordExecutionTime(time.Since(start))
		}
	}()

	// Create context with individual timeout
	ctx, cancel := context.WithTimeout(ctx, eg.timeoutDuration())
	defer cancel()

	// Execute query with retry for transient errors
	return eg.executeWithRetry(ctx, eg.base.GatherSingle, query, t)
}

// GatherRange gathers a range of metrics with timeout management
func (eg *enhancedGatherer) GatherRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]model.MetricSeries, error) {
	queryStart := time.Now()
	defer func() {
		if eg != nil {
			eg.recordExecutionTime(time.Since(queryStart))
		}
	}()

	// Create context with enhanced timeout based on query range size
	adjustedTimeout := eg.calculateRangeTimeout(start, end)
	ctx, cancel := context.WithTimeout(ctx, adjustedTimeout)
	defer cancel()

	// Execute range query with retry
	return eg.executeWithRetryForRange(ctx, query, start, end, step)
}

// executeWithRetry wraps individual queries with retry logic
func (eg *enhancedGatherer) executeWithRetry(
	ctx context.Context,
	queryFunc func(context.Context, model.Query, time.Time) ([]model.MetricSeries, error),
	query string,
	time time.Time,
) ([]model.MetricSeries, error) {
	var result []model.MetricSeries
	var lastErr error
	
	maxRetries := 2
	
	for retry := 0; retry < maxRetries; retry++ {
		if ctx.Err() != nil {
			if eg.metrics != nil {
				eg.markTimeout()
			}
			return nil, fmt.Errorf("query deadline exceeded: %w", ctx.Err())
		}

		result, lastErr = queryFunc(ctx, model.Query{Expr: query}, time)
		if lastErr == nil {
			if eg.metrics != nil {
				eg.markSuccess()
			}
			return result, nil
		}

		// Don't retry context errors
		if isContextError(lastErr) {
			if eg.metrics != nil {
				eg.markTimeout()
			}
			return nil, lastErr
		}

		// Exponential backoff before retry
		if retry < maxRetries-1 {
			backoff := time.Duration(retry+1) * time.Millisecond * 100
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if eg.metrics != nil {
		eg.markFailure()
	}
	return nil, lastErr
}

// executeWithRetryForRange wraps range queries with specific logic
func (eg *enhancedGatherer) executeWithRetryForRange(
	ctx context.Context,
	query string,
	start, end time.Time,
	step time.Duration,
) ([]model.MetricSeries, error) {
	var result []model.MetricSeries
	var lastErr error
	
	maxRetries := 2
	
	for retry := 0; retry < maxRetries; retry++ {
		if ctx.Err() != nil {
			if eg.metrics != nil {
				eg.markTimeout()
			}
			return nil, fmt.Errorf("range query deadline exceeded: %w", ctx.Err())
		}

		result, lastErr = eg.base.GatherRange(ctx, model.Query{Expr: query}, start, end, step)
		if lastErr == nil {
			if eg.metrics != nil {
				eg.markSuccess()
			}
			return result, nil
		}

		// Handle context errors appropriately
		if isContextError(lastErr) {
			if eg.metrics != nil {
				eg.markTimeout()
			}
			return nil, lastErr
		}

		// Backoff for range queries
		if retry < maxRetries-1 {
			backoff := time.Duration(retry+1) * time.Millisecond * 250 // Larger backoff for range queries
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if eg.metrics != nil {
		eg.markFailure()
	}
	return nil, lastErr
}

func (eg *enhancedGatherer) calculateRangeTimeout(start, end time.Time) time.Duration {
	rangeSize := end.Sub(start)
	baseTimeout := eg.timeoutDuration()
	
	// Scale timeout based on range size (longer ranges need more time)
	scaleFactor := float64(rangeSize) / float64(1*time.Hour)
	if scaleFactor > 1 {
		timeout := time.Duration(float64(baseTimeout) * scaleFactor)
		if timeout > 30*time.Second {
			return 30 * time.Second // Cap at 30s for safety
		}
		return timeout
	}
	
	return baseTimeout
}

func (eg *enhancedGatherer) timeoutDuration() time.Duration {
	if eg == nil {
		return DefaultTimeout
	}
	
	eg.mu.RLock()
	defer eg.mu.RUnlock()
	return eg.timeout
}

func (eg *enhancedGatherer) recordExecutionTime(duration time.Duration) {
	if eg.metrics == nil {
		return
	}
	
	eg.metrics.mu.Lock()
	defer eg.metrics.mu.Unlock()
	
	// Moving average calculation for average execution time
	if eg.metrics.queriesSuccessful > 0 {
		oldAvg := float64(eg.metrics.averageExecTime)
		newDuration := float64(duration)
		eg.metrics.averageExecTime = time.Duration((oldAvg + newDuration) / 2)
	} else {
		eg.metrics.averageExecTime = duration
	}
}

func (eg *enhancedGatherer) getLastExecutionTime() time.Duration {
	if eg == nil || eg.metrics == nil {
		return 0
	}
	
	eg.metrics.mu.RLock()
	defer eg.metrics.mu.RUnlock()
	return eg.metrics.averageExecTime
}

func (eg *enhancedGatherer) markSuccess() {
	if eg.metrics == nil {
		return
	}
	
	eg.metrics.mu.Lock()
	defer eg.metrics.mu.Unlock()
	
	eg.metrics.queriesTotal++
	eg.metrics.queriesSuccessful++
}

func (eg *enhancedGatherer) markFailure() {
	if eg.metrics == nil {
		return
	}
	
	eg.metrics.mu.Lock()
	defer eg.metrics.mu.Unlock()
	
	eg.metrics.queriesTotal++
	eg.metrics.queriesFailed++
}

func (eg *enhancedGatherer) markTimeout() {
	if eg.metrics == nil {
		return
	}
	
	eg.metrics.mu.Lock()
	defer eg.metrics.mu.Unlock()
	
	eg.metrics.queriesTotal++
	eg.metrics.queriesTimeout++
}

// isContextError checks if error is context-related
func isContextError(err error) bool {
	return err != nil && (
		strings.Contains(err.Error(), "deadline exceeded") ||
		strings.Contains(err.Error(), "canceled") ||
		strings.Contains(err.Error(), "timeout"))
}

// GetMetrics returns current gatherer statistics
func (eg *enhancedGatherer) GetMetrics() GathererStats {
	if eg.metrics == nil {
		return GathererStats{}
	}
	
	eg.metrics.mu.RLock()
	defer eg.metrics.mu.RUnlock()
	
	return GathererStats{
		TotalQueries:        eg.metrics.queriesTotal,
		SuccessfulQueries:   eg.metrics.queriesSuccessful,
		FailedQueries:       eg.metrics.queriesFailed,
		TimeoutQueries:      eg.metrics.queriesTimeout,
		AverageExecTime:     eg.metrics.averageExecTime,
		LastExecutionTime:   eg.getLastExecutionTime(),
		CurrentTimeout:      eg.timeoutDuration(),
	}
}

// GathererStats contains performance statistics for the gatherer
type GathererStats struct {
	TotalQueries        int64
	SuccessfulQueries   int64
	FailedQueries       int64
	TimeoutQueries      int64
	AverageExecTime     time.Duration
	LastExecutionTime   time.Duration
	CurrentTimeout      time.Duration
}

const (
	DefaultTimeout = 5 * time.Second
	MinTimeout     = 1 * time.Second
	MaxTimeout     = 30 * time.Second
)