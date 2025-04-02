package handlers

import (
	"github.com/kxrxh/social-rating-system/internal/flink"
	"github.com/kxrxh/social-rating-system/internal/repositories"
)

// Dependencies holds shared application dependencies.
type Dependencies struct {
	Repos       map[string]repositories.Repository
	FlinkClient *flink.Client
	// Add other shared dependencies here as the application grows
}
