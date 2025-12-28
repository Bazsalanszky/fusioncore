package gui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/bazsalanszky/fusioncore/assets"
	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/extractor"
	"github.com/bazsalanszky/fusioncore/internal/games"
	"github.com/bazsalanszky/fusioncore/internal/instance"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	"github.com/bazsalanszky/fusioncore/internal/nexus"
	"github.com/bazsalanszky/fusioncore/internal/vfs"
)

type AppState struct {
	mods        []*mod.Mod
	currentGame *games.Game
}

func buildUI(w fyne.Window, state *AppState) (fyne.CanvasObject, *widget.ProgressBar, *widget.Label, *widget.Button, *widget.List) {
	header := newHeader(w)
	usernameLabel, launchButton := newStatusBar(w)
	modList, _ := newModList(w, state)

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// Create main content area with padding
	content := container.NewPadded(
		container.NewBorder(
			header,
			container.NewVBox(
				widget.NewSeparator(),
				container.NewHBox(
					layout.NewSpacer(),
					usernameLabel,
					launchButton,
				),
				progressBar,
			),
			nil,
			nil,
			modList,
		),
	)

	// Menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Add ba2 from file", func() {
			fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					showErrorDialog(err, w)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()

				modsDir, err := state.currentGame.GetModsDir()
				if err != nil {
					showErrorDialog(err, w)
					return
				}

				fileName := reader.URI().Name()
				modName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
				modDir := filepath.Join(modsDir, modName)

				if err := os.MkdirAll(modDir, 0755); err != nil {
					showErrorDialog(err, w)
					return
				}

				destPath := filepath.Join(modDir, fileName)
				destFile, err := os.Create(destPath)
				if err != nil {
					showErrorDialog(err, w)
					return
				}
				defer destFile.Close()

				if _, err := io.Copy(destFile, reader); err != nil {
					showErrorDialog(err, w)
					return
				}

				newMod := &mod.Mod{
					Name:   modName,
					Path:   modDir,
					Active: false,
					ModID:  "local",
					FileID: "local",
					Game:   state.currentGame.ID,
				}
				state.mods = append(state.mods, newMod)
				if err := mod.SaveMods(state.mods, state.currentGame.ID); err != nil {
					showErrorDialog(err, w)
					return
				}
				modList.Refresh()
			}, w)
			fd.SetFilter(storage.NewExtensionFileFilter([]string{state.currentGame.ArchiveExt}))
			fd.Show()
		}),
		fyne.NewMenuItem("Load from URL (nxm://)", func() {
			entry := widget.NewEntry()
			dialog.ShowCustomConfirm("Load from URL", "Load", "Cancel", entry, func(confirm bool) {
				if confirm {
					go handleDownload(entry.Text, progressBar, modList, w, state)
				}
			}, w)
		}),
	)

	accountMenu := fyne.NewMenu("Account",
		fyne.NewMenuItem("Switch Account", func() {
			cfg, err := config.LoadConfig()
			if err != nil {
				showErrorDialog(err, w)
				return
			}
			cfg.APIKey = ""
			if err := config.SaveConfig(cfg); err != nil {
				showErrorDialog(err, w)
				return
			}
			dialog.ShowInformation("Switch Account", "Please restart the application to switch accounts.", w)
		}),
	)

	var gameMenuItems []*fyne.MenuItem
	for _, game := range games.GetSupportedGames() {
		game := game // capture loop variable
		var menuItem *fyne.MenuItem
		menuItem = fyne.NewMenuItem(game.Name, func() {
			if game.ID == state.currentGame.ID {
				return // Already active
			}
			cfg, err := config.LoadConfig()
			if err != nil {
				showErrorDialog(err, w)
				return
			}
			cfg.CurrentGame = game.ID
			if err := config.SaveConfig(cfg); err != nil {
				showErrorDialog(err, w)
				return
			}
			state.currentGame = &game
			mods, err := mod.LoadMods(game.ID)
			if err != nil {
				showErrorDialog(err, w)
				return
			}
			state.mods = mods
			modList.Refresh()
			// Update checkmarks
			for _, item := range gameMenuItems {
				item.Checked = false
			}
			menuItem.Checked = true
		})
		if game.ID == state.currentGame.ID {
			menuItem.Checked = true
		}
		gameMenuItems = append(gameMenuItems, menuItem)
	}
	gamesMenu := fyne.NewMenu("Games", gameMenuItems...)

	mainMenu := fyne.NewMainMenu(fileMenu, accountMenu, gamesMenu)
	w.SetMainMenu(mainMenu)

	return content, progressBar, usernameLabel, launchButton, modList
}

func handleDownload(nxmURL string, progressBar *widget.ProgressBar, modList *widget.List, w fyne.Window, state *AppState) {
	progressBar.Show()
	defer progressBar.Hide()

	info, err := nexus.ParseNxmURL(nxmURL)
	if err != nil {
		showErrorDialog(err, w)
		return
	}

	// Validate game compatibility
	if info.Game != state.currentGame.NexusName {
		dialog.ShowError(fmt.Errorf("mod is for %s but current game is %s", info.Game, state.currentGame.Name), w)
		return
	}

	for i, m := range state.mods {
		if m.ModID == info.ModID {
			if m.FileID == info.FileID {
				dialog.ShowInformation("Mod already installed", "This mod is already installed.", w)
				return
			}

			// A different version of the same mod is already installed, so we ask the user if they want to update.
			dialog.ShowConfirm("Update mod?", fmt.Sprintf("A different version of %s is already installed. Do you want to update?", m.Name), func(ok bool) {
				if !ok {
					return
				}

				if err := vfs.Deactivate(m.Name); err != nil {
					showErrorDialog(err, w)
					return
				}
				if err := os.RemoveAll(m.Path); err != nil {
					showErrorDialog(err, w)
					return
				}
				state.mods = append(state.mods[:i], state.mods[i+1:]...)
			}, w)
		}
	}
	apiKey := os.Getenv("NEXUS_API_KEY")
	if apiKey == "" {
		cfg, err := config.LoadConfig()
		if err != nil {
			showErrorDialog(err, w)
			return
		}
		apiKey = cfg.APIKey
	}

	if _, err := nexus.ValidateAPIKey(apiKey); err != nil {
		showErrorDialog(err, w)
		return
	}

	downloadURL, err := nexus.GetDownloadURL(info, apiKey)
	if err != nil {
		showErrorDialog(err, w)
		return
	}

	destDir, err := state.currentGame.GetModsDir()
	if err != nil {
		showErrorDialog(err, w)
		return
	}

	filePath, err := nexus.DownloadFile(downloadURL, destDir, func(progress float64) {
		progressBar.SetValue(progress)
	})
	if err != nil {
		showErrorDialog(err, w)
		return
	}

	extractDir := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	if err := extractor.Extract(filePath, extractDir); err != nil {
		showErrorDialog(err, w)
		return
	}

	// Clean up the downloaded archive
	if err := os.Remove(filePath); err != nil {
		showErrorDialog(err, w)
	}

	newMod := &mod.Mod{
		Name:   filepath.Base(extractDir),
		Path:   extractDir,
		Active: false,
		ModID:  info.ModID,
		FileID: info.FileID,
		Game:   state.currentGame.ID,
	}
	state.mods = append(state.mods, newMod)
	if err := mod.SaveMods(state.mods, state.currentGame.ID); err != nil {
		showErrorDialog(err, w)
		return
	}
	modList.Refresh()
}

func updateUsername(usernameChan chan string, w fyne.Window) {
	cfg, err := config.LoadConfig()
	if err != nil {
		usernameChan <- "Error: Could not load config"
		showErrorDialog(err, w)
		return
	}

	if cfg.APIKey == "" {
		usernameChan <- "API key not set"
		return
	}

	username, err := nexus.ValidateAPIKey(cfg.APIKey)
	if err != nil {
		usernameChan <- "Invalid API key"
		showErrorDialog(err, w)
		return
	}
	usernameChan <- "Logged in as: " + username
}

func Show(nxmURL string) {
	a := app.NewWithID("eu.toldi.fusioncore")
	a.SetIcon(fyne.NewStaticResource("icon.png", assets.Icon))
	a.Settings().SetTheme(&fusionTheme{})
	w := a.NewWindow("Fusion Core ☢️")
	w.Resize(fyne.NewSize(1000, 700))
	w.CenterOnScreen()

	var progressBar *widget.ProgressBar
	var content fyne.CanvasObject
	var usernameLabel *widget.Label
	var launchButton *widget.Button
	var modList *widget.List
	var state *AppState

	cfg, err := config.LoadConfig()
	if err != nil {
		showErrorDialog(err, w)
	}

	currentGame, err := games.GetGameByID(cfg.CurrentGame)
	if err != nil {
		showErrorDialog(err, w)
		return
	}

	mods, err := mod.LoadMods(cfg.CurrentGame)
	if err != nil {
		showErrorDialog(err, w)
	}
	state = &AppState{mods: mods, currentGame: currentGame}

	// Start single instance server
	instance.StartServer(func(url string) {
		w.RequestFocus()
		if progressBar != nil && modList != nil {
			go handleDownload(url, progressBar, modList, w, state)
		}
	})

	usernameChan := make(chan string)

	if cfg.APIKey == "" {
		apiKeyWindow := newAPIKeyWindow(a, func(apiKey string) {
			cfg.APIKey = apiKey
			if err := config.SaveConfig(cfg); err != nil {
				showErrorDialog(err, w)
			}
			currentGame, err := games.GetGameByID(cfg.CurrentGame)
			if err != nil {
				showErrorDialog(err, w)
				return
			}
			mods, err := mod.LoadMods(cfg.CurrentGame)
			if err != nil {
				showErrorDialog(err, w)
			}
			state = &AppState{mods: mods, currentGame: currentGame}
			content, progressBar, usernameLabel, launchButton, modList = buildUI(w, state)
			w.SetContent(content)
			go updateUsername(usernameChan, w)
			// Start username update loop and safe launch handler now that widgets exist
			go func() {
				for username := range usernameChan {
					if usernameLabel != nil {
						usernameLabel.SetText(username)
					}
				}
			}()
			if launchButton != nil {
				launchButton.OnTapped = func() {
					cmd := exec.Command("steam", "steam://rungameid/"+state.currentGame.AppID)
					if err := cmd.Start(); err != nil {
						showErrorDialog(err, w)
					}
				}
			}
			if nxmURL != "" {
				go handleDownload(nxmURL, progressBar, modList, w, state)
			}
		})
		apiKeyWindow.Show()
	} else {
		content, progressBar, usernameLabel, launchButton, modList = buildUI(w, state)
		w.SetContent(content)
		go updateUsername(usernameChan, w)
		// Start username update loop and safe launch handler now that widgets exist
		go func() {
			for username := range usernameChan {
				if usernameLabel != nil {
					usernameLabel.SetText(username)
				}
			}
		}()
		if launchButton != nil {
			launchButton.OnTapped = func() {
				cmd := exec.Command("steam", "steam://rungameid/"+state.currentGame.AppID)
				if err := cmd.Start(); err != nil {
					showErrorDialog(err, w)
				}
			}
		}
		if nxmURL != "" {
			go handleDownload(nxmURL, progressBar, modList, w, state)
		}
	}

	w.ShowAndRun()
}
