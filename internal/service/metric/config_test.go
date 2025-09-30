package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultEnhancedFeaturesConfig(t *testing.T) {
	cfg := DefaultEnhancedFeaturesConfig()

	assert.True(t, cfg.Enabled, "enhanced features should be enabled by default")
	assert.True(t, cfg.EnableCaching, "caching should be enabled by default")
	assert.True(t, cfg.EnableRetry, "retry should be enabled by default")
	assert.Equal(t, int64(100), cfg.CacheSize, "cache size should be 100")
	assert.Equal(t, 30*time.Second, cfg.CacheTTL, "cache TTL should be 30s")
	assert.Equal(t, 3, cfg.MaxRetries, "max retries should be 3")
	assert.Equal(t, 5*time.Second, cfg.QueryTimeout, "query timeout should be 5s")
	assert.Equal(t, 10, cfg.MaxConcurrentQueries, "max concurrent queries should be 10")
}

func TestLegacyConfig(t *testing.T) {
	cfg := LegacyConfig()

	assert.False(t, cfg.Enabled, "enhanced features should be disabled in legacy mode")
	assert.False(t, cfg.EnableCaching, "caching should be disabled in legacy mode")
	assert.False(t, cfg.EnableRetry, "retry should be disabled in legacy mode")
	assert.Equal(t, time.Duration(0), cfg.QueryTimeout, "query timeout should be 0 in legacy mode")
	assert.Equal(t, 0, cfg.MaxConcurrentQueries, "max concurrent queries should be 0 in legacy mode")
}

func TestEnhancedFeaturesConfigValues(t *testing.T) {
	tests := []struct {
		name     string
		config   EnhancedFeaturesConfig
		validate func(t *testing.T, cfg EnhancedFeaturesConfig)
	}{
		{
			name: "Custom configuration with all features enabled",
			config: EnhancedFeaturesConfig{
				Enabled:              true,
				EnableCaching:        true,
				CacheSize:            200,
				CacheTTL:             60 * time.Second,
				EnableRetry:          true,
				MaxRetries:           5,
				QueryTimeout:         10 * time.Second,
				MaxConcurrentQueries: 20,
			},
			validate: func(t *testing.T, cfg EnhancedFeaturesConfig) {
				assert.True(t, cfg.Enabled)
				assert.True(t, cfg.EnableCaching)
				assert.Equal(t, int64(200), cfg.CacheSize)
				assert.Equal(t, 60*time.Second, cfg.CacheTTL)
				assert.True(t, cfg.EnableRetry)
				assert.Equal(t, 5, cfg.MaxRetries)
				assert.Equal(t, 10*time.Second, cfg.QueryTimeout)
				assert.Equal(t, 20, cfg.MaxConcurrentQueries)
			},
		},
		{
			name: "Configuration with caching disabled but retry enabled",
			config: EnhancedFeaturesConfig{
				Enabled:       true,
				EnableCaching: false,
				EnableRetry:   true,
				MaxRetries:    2,
				QueryTimeout:  3 * time.Second,
			},
			validate: func(t *testing.T, cfg EnhancedFeaturesConfig) {
				assert.True(t, cfg.Enabled)
				assert.False(t, cfg.EnableCaching)
				assert.True(t, cfg.EnableRetry)
				assert.Equal(t, 2, cfg.MaxRetries)
				assert.Equal(t, 3*time.Second, cfg.QueryTimeout)
			},
		},
		{
			name: "Configuration with retry disabled but caching enabled",
			config: EnhancedFeaturesConfig{
				Enabled:       true,
				EnableCaching: true,
				CacheSize:     50,
				CacheTTL:      15 * time.Second,
				EnableRetry:   false,
				QueryTimeout:  2 * time.Second,
			},
			validate: func(t *testing.T, cfg EnhancedFeaturesConfig) {
				assert.True(t, cfg.Enabled)
				assert.True(t, cfg.EnableCaching)
				assert.Equal(t, int64(50), cfg.CacheSize)
				assert.Equal(t, 15*time.Second, cfg.CacheTTL)
				assert.False(t, cfg.EnableRetry)
				assert.Equal(t, 2*time.Second, cfg.QueryTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.config)
		})
	}
}

func TestConfigCompatibility(t *testing.T) {
	t.Run("Default config is not legacy", func(t *testing.T) {
		defaultCfg := DefaultEnhancedFeaturesConfig()
		legacyCfg := LegacyConfig()

		assert.NotEqual(t, defaultCfg.Enabled, legacyCfg.Enabled)
		assert.NotEqual(t, defaultCfg.EnableCaching, legacyCfg.EnableCaching)
		assert.NotEqual(t, defaultCfg.EnableRetry, legacyCfg.EnableRetry)
	})

	t.Run("Legacy config disables all enhanced features", func(t *testing.T) {
		cfg := LegacyConfig()

		assert.False(t, cfg.Enabled)
		assert.False(t, cfg.EnableCaching)
		assert.False(t, cfg.EnableRetry)
	})
}