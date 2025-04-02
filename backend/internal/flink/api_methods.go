package flink

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

// --- Job Related Structs ---

// JobOverview represents basic information about a Flink job.
type JobOverview struct {
	JobID     string `json:"jid"`
	Name      string `json:"name"`
	State     string `json:"state"` // CREATED, RUNNING, FAILING, FAILED, CANCELLING, CANCELED, FINISHED, RESTARTING, SUSPENDED, RECONCILING
	StartTime int64  `json:"start-time"`
	EndTime   int64  `json:"end-time"`
	Duration  int64  `json:"duration"`
	Tasks     struct {
		Total    int `json:"total"`
		Running  int `json:"running"`
		Finished int `json:"finished"`
		Canceled int `json:"canceled"`
		Failed   int `json:"failed"`
	} `json:"tasks"`
}

// JobsResponse is the expected response structure for the /jobs/overview endpoint.
type JobsResponse struct {
	Jobs []JobOverview `json:"jobs"`
}

// --- API Methods ---

// GetJobs retrieves an overview of all jobs (running, finished, etc.).
func (c *Client) GetJobs(ctx context.Context) (*JobsResponse, error) {
	endpoint := "/jobs/overview"
	fullURL := c.baseURL.ResolveReference(&url.URL{Path: endpoint})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL.String(), nil)
	if err != nil {
		c.logger.Error("Failed to create Flink GetJobs request", zap.String("url", fullURL.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	c.logger.Debug("Sending request to Flink API", zap.String("url", fullURL.String()))
	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute Flink GetJobs request", zap.String("url", fullURL.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		c.logger.Error("Flink API returned non-OK status for GetJobs",
			zap.String("url", fullURL.String()),
			zap.Int("status_code", res.StatusCode),
		)
		return nil, fmt.Errorf("flink API returned status %d", res.StatusCode)
	}

	var jobsResponse JobsResponse
	if err := json.NewDecoder(res.Body).Decode(&jobsResponse); err != nil {
		c.logger.Error("Failed to decode Flink GetJobs response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	c.logger.Debug("Successfully retrieved job overview from Flink", zap.Int("job_count", len(jobsResponse.Jobs)))
	return &jobsResponse, nil
}

// TODO: Add more methods for other Flink API endpoints as needed:
// - GetJobDetails(jobID string)
// - SubmitJob(jarID string, entryClass string, programArgs []string, parallelism int, savepointPath string)
// - CancelJob(jobID string)
// - StopJob(jobID string, savepointDir string)
// - GetClusterOverview()
// - GetTaskManagerDetails()
// etc.
