package main

import (
	"embed"
	"flag"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var (
	versionFlag  = flag.String("version", "latest", "Minecraft version to download and launch (default: latest release)")
	usernameFlag = flag.String("username", "Player", "Username for offline mode")
	dirFlag      = flag.String("dir", ".minecraft-lite", "Directory to download files into")
	parallelFlag = flag.Int("parallel", 10, "Number of parallel downloads")
)

// func main() {
// 	flag.Parse()

// 	fmt.Println("Minecraft Lite Launcher")
// 	fmt.Println("Running with flags:")
// 	fmt.Printf("  Version: %s\n", *versionFlag)
// 	fmt.Printf("  Username: %s\n", *usernameFlag)
// 	fmt.Printf("  Directory: %s\n", *dirFlag)
// 	fmt.Printf("  Parallel Downloads: %d\n", *parallelFlag)

// 	// Ensure the target directory exists
// 	if err := os.MkdirAll(*dirFlag, 0755); err != nil {
// 		panic(fmt.Sprintf("Failed to create directory %s: %v", *dirFlag, err))
// 	}

// 	versionManifest, err := FetchVersions(*dirFlag)
// 	if err != nil {
// 		panic(err)
// 	}

// 	version, err := versionManifest.Find(*versionFlag)
// 	if err != nil {
// 		panic(err)
// 	}

// 	manifest, err := GetOrFetchManifest(*dirFlag, version)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Manifest: %v\n", manifest.Id)

// 	err = manifest.DownloadAll(*dirFlag, *parallelFlag)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("All downloads completed successfully.")

// 	err = manifest.LaunchClient(*dirFlag, *usernameFlag)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Minecraft client launched successfully.")
// }

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "minecraft-launcher-lite",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		LogLevel: logger.ERROR,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
