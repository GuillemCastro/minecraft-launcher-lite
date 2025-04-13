package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (m *Manifest) DownloadAll(ctx context.Context, targetDir string, concurrency int) error {
	targetDir = filepath.Join(targetDir, m.Id)
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	if concurrency <= 0 {
		concurrency = 10
	}

	type DownloadTask struct {
		URL      string
		Path     string
		Expected int
		Sha1     string
	}

	var tasks []DownloadTask

	// Helper to add a task
	addTask := func(url, relativePath string, size int, sha1 string) {
		tasks = append(tasks, DownloadTask{
			URL:      url,
			Path:     filepath.Join(targetDir, relativePath),
			Expected: size,
			Sha1:     sha1,
		})
	}

	// Add main downloads
	addTask(m.Downloads.Client.Url, "client.jar", m.Downloads.Client.Size, m.Downloads.Client.Sha1)
	addTask(m.Downloads.Server.Url, "server.jar", m.Downloads.Server.Size, m.Downloads.Server.Sha1)
	addTask(m.Downloads.ClientMappings.Url, "client_mappings.txt", m.Downloads.ClientMappings.Size, m.Downloads.ClientMappings.Sha1)
	addTask(m.Downloads.ServerMappings.Url, "server_mappings.txt", m.Downloads.ServerMappings.Size, m.Downloads.ServerMappings.Sha1)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	// Add logging file
	addTask(m.Logging.Client.File.URL, filepath.Join("logging", m.Logging.Client.File.ID), m.Logging.Client.File.Size, m.Logging.Client.File.Sha1)

	// Add libraries
	for _, lib := range m.Libraries {
		addTask(lib.Downloads.Artifact.Url, filepath.Join("libraries", lib.Downloads.Artifact.Path), lib.Downloads.Artifact.Size, lib.Downloads.Artifact.Sha1)
	}

	// Create a worker pool
	taskChan := make(chan DownloadTask)
	var wg sync.WaitGroup

	// Download assets index
	assetIndexPath := filepath.Join("assets", "indexes", m.AssetIndex.ID+".json")
	assetIndexFullPath := filepath.Join(targetDir, assetIndexPath)
	downloadFile(client, m.AssetIndex.URL, assetIndexFullPath, m.AssetIndex.Sha1)

	indexData, err := os.ReadFile(assetIndexFullPath)
	if err != nil {
		return fmt.Errorf("failed to read asset index: %w", err)
	}

	type AssetIndex struct {
		Objects map[string]struct {
			Hash string `json:"hash"`
			Size int    `json:"size"`
		} `json:"objects"`
	}

	var assets AssetIndex
	if err := json.Unmarshal(indexData, &assets); err != nil {
		return fmt.Errorf("failed to unmarshal asset index: %w", err)
	}

	for _, obj := range assets.Objects {
		hash := obj.Hash
		prefix := hash[:2]
		url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", prefix, hash)
		relativePath := filepath.Join("assets", "objects", prefix, hash)
		addTask(url, relativePath, obj.Size, hash)
	}

	runtime.EventsEmit(ctx, "downloadStart", len(tasks))
	runtime.EventsEmit(ctx, "downloadProgress", 0, len(tasks))
	var progress atomic.Uint64

	worker := func() {
		defer wg.Done()
		for task := range taskChan {
			if err := downloadFile(client, task.URL, task.Path, task.Sha1); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to download %s: %v\n", task.URL, err)
			}
			progress := progress.Add(1)
			runtime.EventsEmit(ctx, "downloadProgress", progress, len(tasks))
		}
	}

	// Start workers
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go worker()
	}

	// Enqueue normal tasks
	for _, task := range tasks {
		taskChan <- task
	}

	close(taskChan)

	// Wait for all downloads to finish
	wg.Wait()

	runtime.EventsEmit(ctx, "downloadFinished")

	return nil
}

// downloadFile downloads a file if it doesn't exist or if its SHA1 doesn't match
func downloadFile(client http.Client, url, path, expectedSHA1 string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		// File exists, verify SHA1
		ok, err := verifySHA1(path, expectedSHA1)
		if err == nil && ok {
			// File is fine, skip downloading
			return nil
		}
	}

	// Otherwise, download
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response downloading %s: %s", url, resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}

func verifySHA1(path, expected string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))
	return sum == expected, nil
}
