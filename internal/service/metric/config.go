package metric

import "time"

// EnhancedFeaturesConfig configures the enhanced metric gathering features
type EnhancedFeaturesConfig struct {
	// Enabled controls whether enhanced features are active
	// When false, uses legacy behavior for compatibility
	Enabled bool

	// EnableCaching enables the metric cache
	EnableCaching bool

	// CacheSize is the maximum number of cache entries
	CacheSize int64

	// CacheTTL is how long cache entries remain valid
	CacheTTL time.Duration

	// EnableRetry enables query retry logic with exponential backoff
	EnableRetry bool

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// QueryTimeout is the default timeout for queries
	QueryTimeout time.Duration

	// MaxConcurrentQueries limits parallel query execution
	MaxConcurrentQueries int
}

// DefaultEnhancedFeaturesConfig returns the default configuration
func DefaultEnhancedFeaturesConfig() EnhancedFeaturesConfig {
	return EnhancedFeaturesConfig{
		Enabled:              true,
		EnableCaching:        true,
		CacheSize:            100,
		CacheTTL:             30 * time.Second,
		EnableRetry:          true,
		MaxRetries:           3,
		QueryTimeout:         5 * time.Second,
		MaxConcurrentQueries: 10,
	}
}

// LegacyConfig returns configuration that maintains backward compatibility
func LegacyConfig() EnhancedFeaturesConfig {
	return EnhancedFeaturesConfig{
		Enabled:              false,
		EnableCaching:        false,
		EnableRetry:          false,
		QueryTimeout:         0, // No explicit timeout
		MaxConcurrentQueries: 0, // No limit
	}
}