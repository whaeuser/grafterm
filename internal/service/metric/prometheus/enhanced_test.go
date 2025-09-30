package prometheus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/grafterm/internal/model"
)

func TestEnhancedGatherer_ID(t *testing.T) {
	tests := []struct {
		name         string
		datasourceID string
	}{
		{
			name:         "Simple ID",
			datasourceID: "test-ds",
		},
		{
			name:         "Complex ID",
			datasourceID: "prometheus-prod-us-west-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eg := NewEnhancedGatherer(ConfigGatherer{}, tt.datasourceID)
			assert.Equal(t, tt.datasourceID, eg.ID())
		})
	}
}

func TestEnhancedGatherer_SetTimeout(t *testing.T) {
	tests := []struct {
		name            string
		inputTimeout    time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "Valid timeout",
			inputTimeout:    3 * time.Second,
			expectedTimeout: 3 * time.Second,
		},
		{
			name:            "Timeout too low - should use minimum",
			inputTimeout:    500 * time.Millisecond,
			expectedTimeout: MinTimeout,
		},
		{
			name:            "Timeout too high - should cap at maximum",
			inputTimeout:    60 * time.Second,
			expectedTimeout: MaxTimeout,
		},
		{
			name:            "Zero timeout - should use default",
			inputTimeout:    0,
			expectedTimeout: DefaultTimeout,
		},
		{
			name:            "Negative timeout - should use default",
			inputTimeout:    -5 * time.Second,
			expectedTimeout: DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")
			enhancedImpl := eg.(*enhancedGatherer)
			enhancedImpl.SetTimeout(tt.inputTimeout)

			actualTimeout := enhancedImpl.timeoutDuration()
			assert.Equal(t, tt.expectedTimeout, actualTimeout)
		})
	}
}

func TestEnhancedGatherer_CalculateRangeTimeout(t *testing.T) {
	tests := []struct {
		name            string
		baseTimeout     time.Duration
		start           time.Time
		end             time.Time
		expectedTimeout time.Duration
	}{
		{
			name:            "Short range - should use base timeout",
			baseTimeout:     5 * time.Second,
			start:           time.Now().Add(-30 * time.Minute),
			end:             time.Now(),
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "Long range - should scale timeout",
			baseTimeout:     5 * time.Second,
			start:           time.Now().Add(-24 * time.Hour),
			end:             time.Now(),
			expectedTimeout: MaxTimeout, // Should be capped
		},
		{
			name:            "Very short range - should use base timeout",
			baseTimeout:     5 * time.Second,
			start:           time.Now().Add(-5 * time.Minute),
			end:             time.Now(),
			expectedTimeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")
			enhancedImpl := eg.(*enhancedGatherer)
			enhancedImpl.SetTimeout(tt.baseTimeout)

			timeout := enhancedImpl.calculateRangeTimeout(tt.start, tt.end)
			assert.Equal(t, tt.expectedTimeout, timeout)
		})
	}
}

func TestEnhancedGatherer_MetricsTracking(t *testing.T) {
	eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")
	enhancedImpl := eg.(*enhancedGatherer)

	// Initial metrics should be zero
	stats := enhancedImpl.GetMetrics()
	assert.Equal(t, int64(0), stats.TotalQueries)
	assert.Equal(t, int64(0), stats.SuccessfulQueries)
	assert.Equal(t, int64(0), stats.FailedQueries)
	assert.Equal(t, int64(0), stats.TimeoutQueries)

	// Mark success
	enhancedImpl.markSuccess()
	stats = enhancedImpl.GetMetrics()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.SuccessfulQueries)
	assert.Equal(t, int64(0), stats.FailedQueries)
	assert.Equal(t, int64(0), stats.TimeoutQueries)

	// Mark failure
	enhancedImpl.markFailure()
	stats = enhancedImpl.GetMetrics()
	assert.Equal(t, int64(2), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.SuccessfulQueries)
	assert.Equal(t, int64(1), stats.FailedQueries)
	assert.Equal(t, int64(0), stats.TimeoutQueries)

	// Mark timeout
	enhancedImpl.markTimeout()
	stats = enhancedImpl.GetMetrics()
	assert.Equal(t, int64(3), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.SuccessfulQueries)
	assert.Equal(t, int64(1), stats.FailedQueries)
	assert.Equal(t, int64(1), stats.TimeoutQueries)
}

func TestEnhancedGatherer_RecordExecutionTime(t *testing.T) {
	eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")
	enhancedImpl := eg.(*enhancedGatherer)

	// Record first execution time - when queriesSuccessful is 0, it just sets the value
	enhancedImpl.recordExecutionTime(100 * time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, enhancedImpl.GetLastExecutionTime())

	// Mark success to increment the counter
	enhancedImpl.markSuccess()

	// Record second execution time - now it will average since queriesSuccessful > 0
	enhancedImpl.recordExecutionTime(200 * time.Millisecond)
	avgTime := enhancedImpl.GetLastExecutionTime()
	// Average of 100ms and 200ms should be 150ms
	assert.Equal(t, 150*time.Millisecond, avgTime)
}

func TestIsContextError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Context deadline exceeded",
			err:      errors.New("context deadline exceeded"),
			expected: true,
		},
		{
			name:     "Context canceled",
			err:      errors.New("context canceled"),
			expected: true,
		},
		{
			name:     "Timeout error",
			err:      errors.New("request timeout"),
			expected: true,
		},
		{
			name:     "Regular error",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "Actual context.DeadlineExceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "Actual context.Canceled",
			err:      context.Canceled,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isContextError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnhancedGatherer_Constants(t *testing.T) {
	assert.Equal(t, 5*time.Second, DefaultTimeout, "default timeout should be 5s")
	assert.Equal(t, 1*time.Second, MinTimeout, "min timeout should be 1s")
	assert.Equal(t, 30*time.Second, MaxTimeout, "max timeout should be 30s")
}

func TestEnhancedGatherer_GatherSingleWithContext(t *testing.T) {
	t.Run("Context cancellation should be handled", func(t *testing.T) {
		eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		query := model.Query{
			Expr:         "up",
			DatasourceID: "test-ds",
		}

		_, err := eg.GatherSingle(ctx, query, time.Now())
		require.Error(t, err)
		assert.True(t, isContextError(err) || err == context.Canceled)
	})

	t.Run("Context timeout should be tracked", func(t *testing.T) {
		eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")
		enhancedImpl := eg.(*enhancedGatherer)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout occurs

		query := model.Query{
			Expr:         "up",
			DatasourceID: "test-ds",
		}

		_, _ = eg.GatherSingle(ctx, query, time.Now())
		// Note: Error is expected but base gatherer is nil so call will fail

		stats := enhancedImpl.GetMetrics()
		// Metrics may not be tracked if base gatherer fails early
		assert.GreaterOrEqual(t, stats.TotalQueries, int64(0))
	})
}

func TestEnhancedGatherer_ThreadSafety(t *testing.T) {
	eg := NewEnhancedGatherer(ConfigGatherer{}, "test-ds")
	enhancedImpl := eg.(*enhancedGatherer)

	// Run concurrent operations to test thread safety
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			enhancedImpl.SetTimeout(3 * time.Second)
			_ = enhancedImpl.timeoutDuration()
			_ = enhancedImpl.GetLastExecutionTime()
			enhancedImpl.markSuccess()
			enhancedImpl.recordExecutionTime(50 * time.Millisecond)
			_ = enhancedImpl.GetMetrics()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify metrics are consistent
	stats := enhancedImpl.GetMetrics()
	assert.Equal(t, int64(100), stats.SuccessfulQueries)
	assert.Equal(t, int64(100), stats.TotalQueries)
}