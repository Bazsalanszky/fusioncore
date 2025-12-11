package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/games"
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
			cfg := &config.Config{APIKey: *apiKeyFlag, CurrentGame: "fallout76"}
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
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}
			game, err := games.GetGameByID(cfg.CurrentGame)
			if err != nil {
				log.Fatalf("Failed to get current game: %v", err)
			}
			mods, err := mod.LoadMods(cfg.CurrentGame)
			if err != nil {
				log.Fatalf("Failed to load mods: %v", err)
			}
			if len(mods) == 0 {
				fmt.Printf("No mods installed for %s.\n", game.Name)
				return
			}
			fmt.Printf("Installed mods for %s:\n", game.Name)
			for _, m := range mods {
				status := "inactive"
				if m.Active {
					status = "active"
				}
				fmt.Printf("- %s (%s)\n", m.Name, status)
			}
			return
		case "games":
			fmt.Println("Supported games:")
			for _, game := range games.GetSupportedGames() {
				fmt.Printf("- %s (ID: %s)\n", game.Name, game.ID)
			}
			return
		case "switch-game":
			if len(os.Args) < 3 {
				fmt.Println("Please provide a game ID. Use 'games' command to see available games.")
				return
			}
			gameID := os.Args[2]
			_, err := games.GetGameByID(gameID)
			if err != nil {
				fmt.Printf("Invalid game ID: %s\n", gameID)
				return
			}
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}
			cfg.CurrentGame = gameID
			if err := config.SaveConfig(cfg); err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}
			game, _ := games.GetGameByID(gameID)
			fmt.Printf("Switched to %s\n", game.Name)
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
