package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/gui"
	"github.com/bazsalanszky/fusioncore/internal/instance"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	fos "github.com/bazsalanszky/fusioncore/internal/os"
	"github.com/bazsalanszky/fusioncore/internal/vfs"
)

func main() {
	nxmURL := ""
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "nxm://") {
		nxmURL = os.Args[1]
	}

	// Try to connect to existing instance
	if instance.TryConnect(nxmURL) {
		return
	}

	// No existing instance, start GUI
	if len(os.Args) == 1 || strings.HasPrefix(os.Args[1], "nxm://") {
		gui.Show(nxmURL)
		return
	}

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
}
