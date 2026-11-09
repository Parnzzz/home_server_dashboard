package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"home-server-dashboard/agent/collector"
)

type MetricsRequest struct {
	CPUPercent       float64 `json:"cpu_percent"`
	CPUTEMPC         float64 `json:"cpu_temp_c"`
	MemoryUsedBytes  uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes uint64  `json:"memory_total_bytes"`
	DiskUsedBytes    uint64  `json:"disk_used_bytes"`
	DiskTotalBytes   uint64  `json:"disk_total_bytes"`
}

type APIClient struct {
	baseURL    string
	agentToken string
	httpClient *http.Client
}

func NewAPIClient(baseURL string, agentToken string) *APIClient {
	return &APIClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		agentToken: agentToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) SendMetrics(metrics *collector.SystemMetrics) error {
	payload := MetricsRequest{
		CPUPercent:       metrics.CPUPercent,
		CPUTEMPC:         metrics.CPUTEMPC,
		MemoryUsedBytes:  metrics.MemoryUsedBytes,
		MemoryTotalBytes: metrics.MemoryTotalBytes,
		DiskUsedBytes:    metrics.DiskUsedBytes,
		DiskTotalBytes:   metrics.DiskTotalBytes,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode metrics: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/v1/metrics",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.agentToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf(
			"API returned status %s: %s",
			resp.Status,
			strings.TrimSpace(string(responseBody)),
		)
	}

	return nil
}
