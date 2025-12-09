package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/extractor"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	"github.com/bazsalanszky/fusioncore/internal/nexus"
	fos "github.com/bazsalanszky/fusioncore/internal/os"
	"github.com/bazsalanszky/fusioncore/internal/prefix"
	"github.com/bazsalanszky/fusioncore/internal/vfs"
)

func main() {
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	apiKeyFlag := loginCmd.String("apikey", "", "Your Nexus Mods API key")

	activateCmd := flag.NewFlagSet("activate", flag.ExitOnError)
	activateModName := activateCmd.String("mod", "", "The name of the mod to activate")

	deactivateCmd := flag.NewFlagSet("deactivate", flag.ExitOnError)
	deactivateModName := deactivateCmd.String("mod", "", "The name of the mod to deactivate")

	registerHandler := flag.Bool("register-handler", false, "Register the application as a protocol handler for nxm URLs")

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			loginCmd.Parse(os.Args[2:])
			if *apiKeyFlag == "" {
				fmt.Println("Please provide your API key with the --apikey flag.")
				return
			}
			cfg := &config.Config{APIKey: *apiKeyFlag}
			if err := config.SaveConfig(cfg); err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}
			fmt.Println("API key saved successfully.")
			return
		case "activate":
			activateCmd.Parse(os.Args[2:])
			if *activateModName == "" {
				fmt.Println("Please provide the name of the mod to activate with the --mod flag.")
				return
			}
			if err := vfs.Activate(*activateModName); err != nil {
				log.Fatalf("Failed to activate mod: %v", err)
			}
			fmt.Printf("Mod %s activated successfully.\n", *activateModName)
			return
		case "deactivate":
			deactivateCmd.Parse(os.Args[2:])
			if *deactivateModName == "" {
				fmt.Println("Please provide the name of the mod to deactivate with the --mod flag.")
				return
			}
			if err := vfs.Deactivate(*deactivateModName); err != nil {
				log.Fatalf("Failed to deactivate mod: %v", err)
			}
			fmt.Printf("Mod %s deactivated successfully.\n", *deactivateModName)
			return
		case "list":
			mods, err := mod.LoadMods()
			if err != nil {
				log.Fatalf("Failed to load mods: %v", err)
			}
			if len(mods) == 0 {
				fmt.Println("No mods installed.")
				return
			}
			fmt.Println("Installed mods:")
			for _, m := range mods {
				status := "inactive"
				if m.Active {
					status = "active"
				}
				fmt.Printf("- %s (%s)\n", m.Name, status)
			}
			return
		}
	}

	flag.Parse()

	if *registerHandler {
		if err := fos.RegisterProtocolHandler(); err != nil {
			log.Fatalf("Failed to register protocol handler: %v", err)
		}
		return
	}

	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "nxm://") {
		nxmURL := os.Args[1]
		fmt.Printf("Parsing nxm URL: %s\n", nxmURL)
		info, err := nexus.ParseNxmURL(nxmURL)
		if err != nil {
			log.Fatalf("Failed to parse nxm URL: %v", err)
		}
		fmt.Printf("Parsed NXM URL: %+v\n", info)

		mods, err := mod.LoadMods()
		if err != nil {
			log.Fatalf("Failed to load mods: %v", err)
		}
		for i, m := range mods {
			if m.ModID == info.ModID {
				if m.FileID == info.FileID {
					fmt.Println("Mod is already installed.")
					return
				}

				fmt.Printf("A different version of %s is already installed. Do you want to update? (y/n): ", m.Name)
				var response string
				fmt.Scanln(&response)

				if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
					fmt.Println("Updating mod...")
					if err := vfs.Deactivate(m.Name); err != nil {
						log.Fatalf("Failed to deactivate old mod version: %v", err)
					}
					if err := os.RemoveAll(m.Path); err != nil {
						log.Fatalf("Failed to remove old mod files: %v", err)
					}
					mods = append(mods[:i], mods[i+1:]...)
					break // Continue to download the new version
				} else {
					fmt.Println("Update cancelled.")
					return
				}
			}
		}
		apiKey := os.Getenv("NEXUS_API_KEY")
		if apiKey == "" {
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}
			apiKey = cfg.APIKey
		}

		username, err := nexus.ValidateAPIKey(apiKey)
		if err != nil {
			log.Fatalf("Invalid Nexus API key: %v", err)
		}
		fmt.Printf("Validated API key for user: %s\n", username)

		if apiKey == "" {
			log.Fatal("Nexus API key not set. Please use the login command or set the NEXUS_API_KEY environment variable.")
		}

		downloadURL, err := nexus.GetDownloadURL(info, apiKey)
		if err != nil {
			log.Fatalf("Failed to get download URL: %v", err)
		}
		fmt.Printf("Got download URL: %s\n", downloadURL)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		destDir := filepath.Join(homeDir, "Games", "FusionCore", "Mods", "Fallout76")

		fmt.Printf("Downloading to %s...\n", destDir)
		filePath, err := nexus.DownloadFile(downloadURL, destDir)
		if err != nil {
			log.Fatalf("Failed to download file: %v", err)
		}
		fmt.Printf("File downloaded successfully to %s\n", filePath)

		fmt.Println("Extracting mod...")
		extractDir := strings.TrimSuffix(filePath, filepath.Ext(filePath))
		if err := extractor.Extract(filePath, extractDir); err != nil {
			log.Fatalf("Failed to extract mod: %v", err)
		}
		fmt.Printf("Mod extracted successfully to %s\n", extractDir)

		// Clean up the downloaded archive
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to remove archive: %v", err)
		}

		newMod := &mod.Mod{
			Name:   filepath.Base(extractDir),
			Path:   extractDir,
			Active: true,
			ModID:  info.ModID,
			FileID: info.FileID,
		}
		mods = append(mods, newMod)
		if err := mod.SaveMods(mods); err != nil {
			log.Fatalf("Failed to save mods: %v", err)
		}
		fmt.Println("Mod added to the library and activated.")

		return
	}

	fmt.Println("Welcome to Fusion Core!")

	fmt.Println("Attempting to find Fallout 76 prefix...")
	prefixPath, err := prefix.FindFallout76Prefix()
	if err != nil {
		log.Fatalf("Error finding prefix: %v", err)
	}

	fmt.Printf("Found Fallout 76 prefix at: %s\n", prefixPath)

	fmt.Println("Syncing links...")
	if err := vfs.SyncLinks(); err != nil {
		log.Fatalf("Error syncing links: %v", err)
	}

	fmt.Println("Links synced successfully!")
}
