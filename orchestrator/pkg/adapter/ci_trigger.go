package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CITrigger triggers a GitHub Actions workflow via repository_dispatch.
type CITrigger struct {
	Token string // GitHub PAT with repo scope
}

// TriggerConfig holds CI trigger configuration from module manifest.
type TriggerConfig struct {
	Repo     string // "owner/repo"
	Workflow string // "mantle-test.yaml"
	Event    string // "mantle-test"
}

// TriggerResult holds the outcome of a CI trigger.
type TriggerResult struct {
	RunURL   string
	Status   string // "completed", "in_progress", "failed"
	Duration time.Duration
}

// Trigger dispatches a workflow and optionally waits for completion.
func (t *CITrigger) Trigger(ctx context.Context, cfg TriggerConfig, payload map[string]string, wait bool) (*TriggerResult, error) {
	if t.Token == "" {
		t.Token = os.Getenv("GITHUB_TOKEN")
		if t.Token == "" {
			return nil, fmt.Errorf("GITHUB_TOKEN not set (required for CI trigger mode)")
		}
	}

	start := time.Now()

	// Send repository_dispatch event
	body := map[string]interface{}{
		"event_type":     cfg.Event,
		"client_payload": payload,
	}
	bodyJSON, _ := json.Marshal(body)

	url := fmt.Sprintf("https://api.github.com/repos/%s/dispatches", cfg.Repo)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dispatch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dispatch failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("  Triggered %s/%s (event: %s)\n", cfg.Repo, cfg.Workflow, cfg.Event)

	result := &TriggerResult{
		RunURL: fmt.Sprintf("https://github.com/%s/actions", cfg.Repo),
		Status: "triggered",
	}

	if !wait {
		result.Duration = time.Since(start)
		return result, nil
	}

	// Wait for the workflow run to complete
	fmt.Printf("  Waiting for workflow to complete...\n")
	runID, err := t.findLatestRun(ctx, cfg)
	if err != nil {
		result.Status = "unknown"
		result.Duration = time.Since(start)
		return result, fmt.Errorf("could not find workflow run: %w", err)
	}

	result.RunURL = fmt.Sprintf("https://github.com/%s/actions/runs/%d", cfg.Repo, runID)
	conclusion, err := t.waitForRun(ctx, cfg.Repo, runID)
	if err != nil {
		result.Status = "error"
		result.Duration = time.Since(start)
		return result, err
	}

	result.Status = conclusion
	result.Duration = time.Since(start)
	fmt.Printf("  Workflow completed: %s (%s)\n", conclusion, result.Duration.Round(time.Second))
	return result, nil
}

// findLatestRun finds the most recent workflow run triggered by repository_dispatch.
func (t *CITrigger) findLatestRun(ctx context.Context, cfg TriggerConfig) (int64, error) {
	// Wait a bit for GitHub to register the run
	time.Sleep(5 * time.Second)

	for i := 0; i < 12; i++ { // retry for 60 seconds
		url := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs?event=repository_dispatch&per_page=1", cfg.Repo)
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+t.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		var result struct {
			WorkflowRuns []struct {
				ID     int64  `json:"id"`
				Status string `json:"status"`
			} `json:"workflow_runs"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if len(result.WorkflowRuns) > 0 {
			return result.WorkflowRuns[0].ID, nil
		}
		time.Sleep(5 * time.Second)
	}
	return 0, fmt.Errorf("no workflow run found after 60s")
}

// waitForRun polls a workflow run until it completes.
func (t *CITrigger) waitForRun(ctx context.Context, repo string, runID int64) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d", repo, runID)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+t.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		var run struct {
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		}
		json.NewDecoder(resp.Body).Decode(&run)
		resp.Body.Close()

		if run.Status == "completed" {
			return run.Conclusion, nil
		}

		time.Sleep(15 * time.Second)
	}
}
