package main

import (
	"fmt"
	"home-server-dashboard/agent/client"
	"home-server-dashboard/agent/collector"
	"log"
	"os"
	"time"
)

func main() {
	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8080"
	}

	agentToken := os.Getenv("AGENT_TOKEN")
	if agentToken == "" {
		log.Fatal("AGENT_TOKEN is not configured")
	}

	apiClient := client.NewAPIClient(apiBaseURL, agentToken)

	fmt.Println("===== Home Server Agent =====")
	fmt.Println("Monitoring started... Press Ctrl+C to stop")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		metrics, err := collector.Collect()
		if err != nil {
			log.Println("Failed to collect metrics:", err)
		} else {
			err := apiClient.SendMetrics(metrics)

			if err != nil {
				log.Println("Failed to send metrics:", err)
			} else {
				fmt.Printf(
					"Metrics sent | CPU: %.2f%% | CPUTEMPC: %.2f%% | RAM: %.2f / %.2f GB | Disk: %.2f / %.2f GB\n",
					metrics.CPUPercent,
					metrics.CPUTEMPC,
					float64(metrics.MemoryUsedBytes)/1024/1024/1024,
					float64(metrics.MemoryTotalBytes)/1024/1024/1024,
					float64(metrics.DiskUsedBytes)/1024/1024/1024,
					float64(metrics.DiskTotalBytes)/1024/1024/1024,
				)
			}
		}

		<-ticker.C
	}
}
