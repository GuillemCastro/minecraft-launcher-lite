package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (m *Manifest) LaunchClient(ctx context.Context, targetDir, username string) error {
	targetDir = filepath.Join(targetDir, m.Id)

	clientJar := filepath.Join(targetDir, "client.jar")
	if _, err := os.Stat(clientJar); err != nil {
		return fmt.Errorf("client.jar not found: %w", err)
	}

	javaPath, err := exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("java not found in PATH: %w", err)
	}

	// Build the classpath: all libraries + client.jar
	var classpath []string

	// Add libraries
	for _, lib := range m.Libraries {
		if lib.Downloads.Artifact.Path == "" {
			continue // some libraries are "natives" or have classifiers only
		}
		libPath := filepath.Join(targetDir, "libraries", lib.Downloads.Artifact.Path)
		if _, err := os.Stat(libPath); err == nil {
			classpath = append(classpath, libPath)
		}
	}

	// Add client.jar at the end
	classpath = append(classpath, clientJar)

	assetsDir := filepath.Join(targetDir, "assets")
	gameDir := filepath.Join(targetDir, "game")

	// Make sure gameDir exists
	if err := os.MkdirAll(gameDir, 0755); err != nil {
		return fmt.Errorf("failed to create game dir: %w", err)
	}

	// Build arguments
	args := []string{
		"-cp", strings.Join(classpath, string(os.PathListSeparator)),
		"net.minecraft.client.main.Main",
		"--username", username,
		"--version", m.Id,
		"--gameDir", gameDir,
		"--assetsDir", assetsDir,
		"--assetIndex", m.AssetIndex.ID,
		"--userType", "legacy",
		"--accessToken", "0",
		"--uuid", RandomUUID(),
	}

	cmd := exec.Command(javaPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("Launching Minecraft...")

	runtime.EventsEmit(ctx, "launching", m.Id)
	return cmd.Run()
}

func RandomUUID() string {
	// Generate a random UUID (v4)
	return uuid.New().String()
}
