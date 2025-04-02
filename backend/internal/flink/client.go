package flink

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kxrxh/social-rating-system/internal/config"
	"github.com/kxrxh/social-rating-system/internal/logger"
	"go.uber.org/zap"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new Flink API client.
func NewClient(cfg config.FlinkConfig) (*Client, error) {
	baseURLString := fmt.Sprintf("http://%s:%d", cfg.JobManagerHost, cfg.RestPort)
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		logger.Log.Error("Invalid Flink base URL", zap.String("url", baseURLString), zap.Error(err))
		return nil, fmt.Errorf("invalid flink base URL '%s': %w", baseURLString, err)
	}

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Default timeout, can be configurable
		},
		logger: logger.Log.Named("flink_client"),
	}

	client.logger.Info("Flink client initialized", zap.String("base_url", baseURL.String()))

	pingURL := baseURL.ResolveReference(&url.URL{Path: "/ping"})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, pingURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ping request: %w", err)
	}

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ping request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ping request returned status %d", res.StatusCode)
	}

	client.logger.Info("Flink client ping successful")

	return client, nil
}
