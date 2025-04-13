package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []Version `json:"versions"`
}

type Version struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	Url         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
}

const cacheFile = "version_manifest.json"
const cacheDuration = time.Hour // Cache duration of 1 hour

// FetchVersions fetches the version manifest either from cache or downloads it.
func FetchVersions(dir string) (*VersionManifest, error) {
	cacheFile := filepath.Join(dir, cacheFile)
	// Check if the cache exists
	if _, err := os.Stat(cacheFile); err == nil {
		// Cache exists, check if it's expired
		fileInfo, err := os.Stat(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("failed to stat cache file: %v", err)
		}

		// If the cache file is too old, download a new one
		if time.Since(fileInfo.ModTime()) < cacheDuration {
			// Cache is still fresh, load it
			fmt.Println("Using cached version manifest.")
			return loadVersionManifestFromFile(cacheFile)
		}
	}

	// Cache doesn't exist or is expired, download a new one
	fmt.Println("Downloading version manifest.")
	resp, err := http.Get("https://launchermeta.mojang.com/mc/game/version_manifest.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version manifest: %v", err)
	}
	defer resp.Body.Close()

	var versionManifest VersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&versionManifest); err != nil {
		return nil, fmt.Errorf("failed to decode version manifest: %v", err)
	}

	// Save the fresh version manifest to cache
	if err := saveVersionManifestToFile(cacheFile, &versionManifest); err != nil {
		return nil, fmt.Errorf("failed to save version manifest to cache: %v", err)
	}

	return &versionManifest, nil
}

// loadVersionManifestFromFile loads the version manifest from a local file.
func loadVersionManifestFromFile(filePath string) (*VersionManifest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open version manifest file: %v", err)
	}
	defer file.Close()

	var versionManifest VersionManifest
	if err := json.NewDecoder(file).Decode(&versionManifest); err != nil {
		return nil, fmt.Errorf("failed to decode version manifest from file: %v", err)
	}

	return &versionManifest, nil
}

// saveVersionManifestToFile saves the version manifest to a local file.
func saveVersionManifestToFile(filePath string, versionManifest *VersionManifest) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create version manifest file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(versionManifest); err != nil {
		return fmt.Errorf("failed to encode version manifest to file: %v", err)
	}

	return nil
}

func (vm *VersionManifest) Find(id string) (*Version, error) {
	if id == "latest" {
		id = vm.Latest.Release
	}
	for _, v := range vm.Versions {
		if v.Id == id {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("version %s not found", id)
}
