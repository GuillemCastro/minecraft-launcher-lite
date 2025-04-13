package main

import (
	"context"
	"fmt"
	"os"
)

const (
	DEFAULT_TARGET_DIR = ".minecraft-lite"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// ensure DEFAULT_TARGET_DIR exists
	if err := os.MkdirAll(DEFAULT_TARGET_DIR, 0755); err != nil {
		fmt.Println("Failed to create directory:", err)
		panic(err)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetVersions() []string {
	versions, err := FetchVersions(DEFAULT_TARGET_DIR)
	if err != nil {
		fmt.Println("Error fetching versions:", err)
		return nil
	}
	var versionList []string
	for _, version := range versions.Versions {
		if version.Type != "release" {
			continue
		}
		versionList = append(versionList, version.Id)
	}
	return versionList
}

func (a *App) Launch(version string, username string) {
	versions, err := FetchVersions(DEFAULT_TARGET_DIR)
	if err != nil {
		fmt.Println("Error fetching versions:", err)
		return
	}
	versionInfo, err := versions.Find(version)
	if err != nil {
		fmt.Println("Error finding version:", err)
		return
	}

	manifest, err := GetOrFetchManifest(DEFAULT_TARGET_DIR, versionInfo)
	if err != nil {
		fmt.Println("Error fetching manifest:", err)
		return
	}
	err = manifest.DownloadAll(a.ctx, DEFAULT_TARGET_DIR, 10)
	if err != nil {
		fmt.Println("Error downloading files:", err)
		return
	}
	err = manifest.LaunchClient(a.ctx, DEFAULT_TARGET_DIR, username)
	if err != nil {
		fmt.Println("Error launching client:", err)
		return
	}
	fmt.Println("Minecraft client launched successfully.")
}
