package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Manifest struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	JavaVersion struct {
		Component    string `json:"component"`
		MajorVersion int    `json:"majorVersion"`
	} `json:"javaVersion"`
	MainClass              string `json:"mainClass"`
	MinimumLauncherVersion int32  `json:"minimumLauncherVersion"`
	ReleaseTime            string `json:"releaseTime"`
	Time                   string `json:"time"`
	Assets                 string `json:"assets"`
	ComplianceLevel        int    `json:"complianceLevel"`
	Downloads              struct {
		Client         Download `json:"client"`
		Server         Download `json:"server"`
		ClientMappings Download `json:"client_mappings"`
		ServerMappings Download `json:"server_mappings"`
	} `json:"downloads"`
	AssetIndex AssetIndex `json:"assetIndex"`
	Libraries  []Library  `json:"libraries"`
	Logging    struct {
		Client LoggingClient `json:"client"`
	} `json:"logging"`
	Arguments struct {
		Game []Argument `json:"game"`
		Jvm  []Argument `json:"jvm"`
	} `json:"arguments"`
}

type Argument struct {
	Rules []Rule      `json:"rules,omitempty"`
	Value interface{} `json:"value,omitempty"` // string or []string
}

// Custom unmarshaler for Argument
func (a *Argument) UnmarshalJSON(data []byte) error {
	// Check if it's a simple string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		a.Value = str
		return nil
	}

	// Otherwise unmarshal as full struct
	type Alias Argument
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	*a = Argument(alias)
	return nil
}

type Library struct {
	Name      string `json:"name"`
	Downloads struct {
		Artifact Download `json:"artifact"`
	} `json:"downloads"`
}

type Download struct {
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	Url  string `json:"url"`
	Path string `json:"path,omitempty"`
}

type Rule struct {
	Action   string    `json:"action"`
	Features *Features `json:"features,omitempty"`
	OS       *OS       `json:"os,omitempty"`
}

type Features struct {
	IsDemoUser              *bool `json:"is_demo_user,omitempty"`
	HasCustomResolution     *bool `json:"has_custom_resolution,omitempty"`
	HasQuickPlaysSupport    *bool `json:"has_quick_plays_support,omitempty"`
	IsQuickPlaySingleplayer *bool `json:"is_quick_play_singleplayer,omitempty"`
	IsQuickPlayMultiplayer  *bool `json:"is_quick_play_multiplayer,omitempty"`
	IsQuickPlayRealms       *bool `json:"is_quick_play_realms,omitempty"`
}

type OS struct {
	Name string `json:"name,omitempty"`
	Arch string `json:"arch,omitempty"`
}

type AssetIndex struct {
	ID        string `json:"id"`
	Sha1      string `json:"sha1"`
	Size      int    `json:"size"`
	TotalSize int    `json:"totalSize"`
	URL       string `json:"url"`
}

type LoggingClient struct {
	Argument string      `json:"argument"`
	File     LoggingFile `json:"file"`
	Type     string      `json:"type"`
}

type LoggingFile struct {
	ID   string `json:"id"`
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

func GetOrFetchManifest(targetDir string, version *Version) (*Manifest, error) {
	manifest, err := FindManifestById(*dirFlag, version.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest by ID: %w", err)
	}
	if manifest == nil {
		manifest, err = FetchManifest(*dirFlag, version.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch manifest: %w", err)
		}
	}
	return manifest, nil
}

func FindManifestById(targetDir string, id string) (*Manifest, error) {
	manifestPath := filepath.Join(targetDir, id, id+".json")
	// Check if the manifest file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, nil // Manifest not found
	}

	// Read the manifest file
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer file.Close()

	// Unmarshal the JSON data into a Manifest struct
	var manifest Manifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest file: %w", err)
	}
	return &manifest, nil
}

func FetchManifest(targetDir string, url string) (*Manifest, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read version manifest body: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// Save the manifest to the target directory
	manifestPath := filepath.Join(targetDir, manifest.Id, manifest.Id+".json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create manifest directory: %w", err)
	}
	file, err := os.Create(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer file.Close()
	// Write the manifest to the file
	if _, err := file.Write(body); err != nil {
		return nil, fmt.Errorf("failed to write manifest file: %w", err)
	}

	return &manifest, nil
}
