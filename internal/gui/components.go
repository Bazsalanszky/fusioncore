package gui

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bazsalanszky/fusioncore/internal/config"
	"github.com/bazsalanszky/fusioncore/internal/games"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	"github.com/bazsalanszky/fusioncore/internal/vfs"
	"gopkg.in/ini.v1"
)

func newAPIKeyWindow(a fyne.App, onSave func(string)) fyne.Window {
	w := a.NewWindow("Nexus API Key Setup")
	w.Resize(fyne.NewSize(500, 300))
	w.CenterOnScreen()

	// Header
	header := widget.NewRichTextFromMarkdown("## 🔑 Nexus API Key Required")

	// Info card
	infoCard := canvas.NewRectangle(color.NRGBA{R: 33, G: 150, B: 243, A: 50})
	infoCard.CornerRadius = 8

	infoText := widget.NewRichTextFromMarkdown(
		"To download mods from Nexus, you need to provide your API key.\n\n" +
			"**Steps:**\n" +
			"1. Click the link below to open Nexus Mods\n" +
			"2. Log in to your account\n" +
			"3. Copy your Personal API Key\n" +
			"4. Paste it in the field below",
	)

	infoContainer := container.NewStack(
		infoCard,
		container.NewPadded(infoText),
	)

	nexusURL, _ := url.Parse("https://www.nexusmods.com/users/myaccount?tab=api")
	hyperlink := widget.NewHyperlink("🌐 Get Your API Key", nexusURL)

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Paste your API key here...")

	saveButton := widget.NewButtonWithIcon("Save & Continue", theme.ConfirmIcon(), func() {
		onSave(apiKeyEntry.Text)
		w.Close()
	})
	saveButton.Importance = widget.HighImportance

	cancelButton := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		w.Close()
		a.Quit()
	})

	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		cancelButton,
		saveButton,
	)

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		infoContainer,
		hyperlink,
		widget.NewLabel("API Key:"),
		apiKeyEntry,
		buttonRow,
	)

	w.SetContent(container.NewPadded(content))
	w.SetOnClosed(func() {
		// If window is closed without saving, quit the app
		if apiKeyEntry.Text == "" {
			a.Quit()
		}
	})
	return w
}

func showErrorDialog(err error, w fyne.Window) {
	dialog.ShowError(err, w)
}

// Custom theme for Fusion Core
type fusionTheme struct{}

func (f *fusionTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255}
	case theme.ColorNameButton:
		return color.NRGBA{R: 63, G: 81, B: 181, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 255, G: 193, B: 7, A: 255} // Fallout yellow
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 244, G: 67, B: 54, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (f *fusionTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (f *fusionTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (f *fusionTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func newModList(w fyne.Window, state *AppState) (*widget.List, []*mod.Mod) {
	var modList *widget.List
	modList = widget.NewList(
		func() int {
			return len(state.mods)
		},
		func() fyne.CanvasObject {
			// Create mod card
			card := canvas.NewRectangle(color.NRGBA{R: 40, G: 44, B: 52, A: 255})
			card.CornerRadius = 8

			statusIndicator := canvas.NewCircle(color.NRGBA{R: 100, G: 100, B: 100, A: 255})
			statusIndicator.Resize(fyne.NewSize(12, 12))

			modName := widget.NewRichTextFromMarkdown("**Mod Name**")
			modStatus := widget.NewLabel("Status")
			modStatus.TextStyle.Italic = true

			upBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), nil)
			downBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), nil)
			activateBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil)
			activateBtn.Importance = widget.HighImportance
			deactivateBtn := widget.NewButtonWithIcon("", theme.MediaPauseIcon(), nil)
			uninstallBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			uninstallBtn.Importance = widget.DangerImportance

			headerRow := container.NewHBox(
				statusIndicator,
				modName,
				layout.NewSpacer(),
				upBtn,
				downBtn,
				activateBtn,
				deactivateBtn,
				uninstallBtn,
			)

			cardContent := container.NewVBox(
				headerRow,
				modStatus,
			)

			return container.NewStack(
				card,
				container.NewPadded(cardContent),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			m := state.mods[i]
			stackContainer := o.(*fyne.Container)
			cardContent := stackContainer.Objects[1].(*fyne.Container)
			vboxContent := cardContent.Objects[0].(*fyne.Container)
			headerRow := vboxContent.Objects[0].(*fyne.Container)
			statusLabel := vboxContent.Objects[1].(*widget.Label)

			statusIndicator := headerRow.Objects[0].(*canvas.Circle)
			modName := headerRow.Objects[1].(*widget.RichText)
			upBtn := headerRow.Objects[3].(*widget.Button)
			downBtn := headerRow.Objects[4].(*widget.Button)
			activateBtn := headerRow.Objects[5].(*widget.Button)
			deactivateBtn := headerRow.Objects[6].(*widget.Button)
			uninstallBtn := headerRow.Objects[7].(*widget.Button)

			modName.ParseMarkdown(fmt.Sprintf("**%s**", m.Name))

			if m.Active {
				statusIndicator.FillColor = color.NRGBA{R: 76, G: 175, B: 80, A: 255} // Green
				statusLabel.SetText("✓ Active")
				statusLabel.Importance = widget.SuccessImportance
			} else {
				statusIndicator.FillColor = color.NRGBA{R: 158, G: 158, B: 158, A: 255} // Gray
				statusLabel.SetText("⏸ Inactive")
				statusLabel.Importance = widget.MediumImportance
			}
			statusIndicator.Refresh()

			// Move up button
			upBtn.OnTapped = func() {
				if i > 0 {
					state.mods[i], state.mods[i-1] = state.mods[i-1], state.mods[i]
					if err := mod.SaveMods(state.mods, state.currentGame.ID); err != nil {
						showErrorDialog(err, w)
						return
					}
					if err := updateLoadOrder(state.mods, state.currentGame); err != nil {
						showErrorDialog(err, w)
					}
					modList.Refresh()
				}
			}

			// Move down button
			downBtn.OnTapped = func() {
				if i < len(state.mods)-1 {
					state.mods[i], state.mods[i+1] = state.mods[i+1], state.mods[i]
					if err := mod.SaveMods(state.mods, state.currentGame.ID); err != nil {
						showErrorDialog(err, w)
						return
					}
					if err := updateLoadOrder(state.mods, state.currentGame); err != nil {
						showErrorDialog(err, w)
					}
					modList.Refresh()
				}
			}

			// Disable buttons at boundaries
			upBtn.Enable()
			downBtn.Enable()
			if i == 0 {
				upBtn.Disable()
			}
			if i == len(state.mods)-1 {
				downBtn.Disable()
			}

			activateBtn.OnTapped = func() {
				if err := vfs.Activate(m.Name); err != nil {
					showErrorDialog(err, w)
					return
				}
				state.mods, _ = mod.LoadMods(state.currentGame.ID)
				modList.Refresh()
			}
			deactivateBtn.OnTapped = func() {
				if err := vfs.Deactivate(m.Name); err != nil {
					showErrorDialog(err, w)
					return
				}
				state.mods, _ = mod.LoadMods(state.currentGame.ID)
				modList.Refresh()
			}
			uninstallBtn.OnTapped = func() {
				dialog.ShowConfirm("Uninstall Mod", "Are you sure you want to uninstall "+m.Name+"?", func(confirm bool) {
					if !confirm {
						return
					}

					if m.Active {
						if err := vfs.Deactivate(m.Name); err != nil {
							showErrorDialog(err, w)
						}
					}

					if err := os.RemoveAll(m.Path); err != nil {
						showErrorDialog(err, w)
					}

					var newMods []*mod.Mod
					for _, modEntry := range state.mods {
						if modEntry.Name != m.Name {
							newMods = append(newMods, modEntry)
						}
					}
					mod.SaveMods(newMods, state.currentGame.ID)
					state.mods = newMods
					modList.Refresh()
				}, w)
			}

			if m.Active {
				activateBtn.Disable()
				deactivateBtn.Enable()
			} else {
				activateBtn.Enable()
				deactivateBtn.Disable()
			}
		},
	)

	return modList, state.mods
}

func newHeader(w fyne.Window) fyne.CanvasObject {
	title := widget.NewRichTextFromMarkdown("# Fusion Core ☢️")
	subtitle := widget.NewLabel("Powering Fallout Mods on Linux")
	subtitle.TextStyle.Italic = true

	header := container.NewVBox(
		title,
		subtitle,
	)

	return container.NewPadded(header)
}

func newStatusBar(w fyne.Window) (*widget.Label, *widget.Button) {
	usernameLabel := widget.NewLabel("Fetching username...")
	usernameLabel.TextStyle.Monospace = true

	launchButton := widget.NewButtonWithIcon("Launch Game", theme.MediaPlayIcon(), nil)
	launchButton.Importance = widget.HighImportance

	return usernameLabel, launchButton
}

// updateLoadOrder updates the INI file with the current mod order
func updateLoadOrder(mods []*mod.Mod, game *games.Game) error {
	var archives []string
	for _, m := range mods {
		if m.Active {
			archiveFiles, err := findArchiveFiles(m.Path, game.ArchiveExt)
			if err != nil {
				return err
			}
			archives = append(archives, archiveFiles...)
		}
	}
	return setArchiveList(archives, game)
}

// findArchiveFiles finds all archive files with the given extension in a directory
func findArchiveFiles(dir, ext string) ([]string, error) {
	var archiveFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ext {
			archiveFiles = append(archiveFiles, info.Name())
		}
		return nil
	})
	return archiveFiles, err
}

// setArchiveList sets the complete archive list in the INI file
func setArchiveList(archives []string, game *games.Game) error {
	prefixPath, err := game.FindCompatdata()
	if err != nil {
		return err
	}

	iniPath := config.GetCustomIniPath(prefixPath, game.ConfigFile)
	cfg, err := ini.Load(iniPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = ini.Empty()
		} else {
			return fmt.Errorf("failed to load %s: %w", game.ConfigFile, err)
		}
	}

	section := cfg.Section("Archive")
	key := section.Key("sResourceArchive2List")
	key.SetValue(strings.Join(archives, ", "))

	return cfg.SaveTo(iniPath)
}

func newSettingsWindow(a fyne.App, w fyne.Window, state *AppState) fyne.Window {
	settingsWindow := a.NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(800, 600))
	settingsWindow.CenterOnScreen()

	cfg, err := config.LoadConfig()
	if err != nil {
		showErrorDialog(err, w)
		return settingsWindow
	}

	// Create forms for game paths and compatdata paths
	var gamePathItems []*widget.FormItem
	var compatdataPathItems []*widget.FormItem

	for _, game := range games.GetSupportedGames() {
		game := game // capture loop variable

		// ===== GAME DIRECTORY SETTINGS =====
		// Get current game path (custom or auto-discovered)
		currentGamePath := ""
		if cfg.GamePaths != nil {
			if customPath, ok := cfg.GamePaths[game.ID]; ok && customPath != "" {
				currentGamePath = customPath
			}
		}
		if currentGamePath == "" {
			// Try to discover
			if discoveredPath, err := game.FindGameDir(); err == nil {
				currentGamePath = discoveredPath
			} else {
				currentGamePath = "Not found"
			}
		}

		gamePathLabel := widget.NewLabel(currentGamePath)
		gamePathLabel.Wrapping = fyne.TextTruncate

		gameBrowseButton := widget.NewButton("Browse", func() {
			dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
				if err != nil {
					showErrorDialog(err, settingsWindow)
					return
				}
				if uri == nil {
					return
				}

				selectedPath := uri.Path()

				// Verify this looks like a valid game directory
				dataDir := filepath.Join(selectedPath, game.DataSubDir)
				if _, err := os.Stat(dataDir); err != nil {
					dialog.ShowError(fmt.Errorf("Selected directory does not contain a '%s' subdirectory. Please select the game's root directory", game.DataSubDir), settingsWindow)
					return
				}

				// Save the custom path
				cfg.GamePaths[game.ID] = selectedPath
				if err := config.SaveConfig(cfg); err != nil {
					showErrorDialog(err, settingsWindow)
					return
				}

				gamePathLabel.SetText(selectedPath)
				dialog.ShowInformation("Success", fmt.Sprintf("%s game path updated successfully", game.Name), settingsWindow)
			}, settingsWindow)
		})

		gameClearButton := widget.NewButton("Clear", func() {
			if cfg.GamePaths != nil {
				delete(cfg.GamePaths, game.ID)
				if err := config.SaveConfig(cfg); err != nil {
					showErrorDialog(err, settingsWindow)
					return
				}

				// Try to auto-discover again
				if discoveredPath, err := game.FindGameDir(); err == nil {
					gamePathLabel.SetText(discoveredPath)
				} else {
					gamePathLabel.SetText("Not found")
				}
				dialog.ShowInformation("Success", fmt.Sprintf("%s custom game path cleared. Using auto-discovery", game.Name), settingsWindow)
			}
		})

		gameButtonsContainer := container.NewHBox(gameBrowseButton, gameClearButton)
		gamePathContainer := container.NewBorder(nil, nil, nil, gameButtonsContainer, gamePathLabel)
		gamePathItems = append(gamePathItems, widget.NewFormItem(game.Name, gamePathContainer))

		// ===== COMPATDATA DIRECTORY SETTINGS =====
		// Get current compatdata path (custom or auto-discovered)
		currentCompatdataPath := ""
		if cfg.CompatdataPaths != nil {
			if customPath, ok := cfg.CompatdataPaths[game.ID]; ok && customPath != "" {
				currentCompatdataPath = customPath
			}
		}
		if currentCompatdataPath == "" {
			// Try to discover
			if discoveredPath, err := game.FindCompatdata(); err == nil {
				currentCompatdataPath = discoveredPath
			} else {
				currentCompatdataPath = "Not found"
			}
		}

		compatdataPathLabel := widget.NewLabel(currentCompatdataPath)
		compatdataPathLabel.Wrapping = fyne.TextTruncate

		compatdataBrowseButton := widget.NewButton("Browse", func() {
			dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
				if err != nil {
					showErrorDialog(err, settingsWindow)
					return
				}
				if uri == nil {
					return
				}

				selectedPath := uri.Path()

				// Verify this looks like a valid compatdata directory (should have pfx subdirectory)
				pfxDir := filepath.Join(selectedPath, "pfx")
				if _, err := os.Stat(pfxDir); err != nil {
					dialog.ShowError(fmt.Errorf("Selected directory does not contain a 'pfx' subdirectory. Please select the compatdata directory (e.g., steamapps/compatdata/%s)", game.AppID), settingsWindow)
					return
				}

				// Save the custom path
				cfg.CompatdataPaths[game.ID] = selectedPath
				if err := config.SaveConfig(cfg); err != nil {
					showErrorDialog(err, settingsWindow)
					return
				}

				compatdataPathLabel.SetText(selectedPath)
				dialog.ShowInformation("Success", fmt.Sprintf("%s compatdata path updated successfully", game.Name), settingsWindow)
			}, settingsWindow)
		})

		compatdataClearButton := widget.NewButton("Clear", func() {
			if cfg.CompatdataPaths != nil {
				delete(cfg.CompatdataPaths, game.ID)
				if err := config.SaveConfig(cfg); err != nil {
					showErrorDialog(err, settingsWindow)
					return
				}

				// Try to auto-discover again
				if discoveredPath, err := game.FindCompatdata(); err == nil {
					compatdataPathLabel.SetText(discoveredPath)
				} else {
					compatdataPathLabel.SetText("Not found")
				}
				dialog.ShowInformation("Success", fmt.Sprintf("%s custom compatdata path cleared. Using auto-discovery", game.Name), settingsWindow)
			}
		})

		compatdataButtonsContainer := container.NewHBox(compatdataBrowseButton, compatdataClearButton)
		compatdataPathContainer := container.NewBorder(nil, nil, nil, compatdataButtonsContainer, compatdataPathLabel)
		compatdataPathItems = append(compatdataPathItems, widget.NewFormItem(game.Name, compatdataPathContainer))
	}

	gameForm := widget.NewForm(gamePathItems...)
	compatdataForm := widget.NewForm(compatdataPathItems...)

	// Headers
	gameHeader := widget.NewRichTextFromMarkdown("### Game Installation Paths")
	gameInfoText := widget.NewLabel("Configure the game's root directory (where the executable is located).")
	gameInfoText.Wrapping = fyne.TextWrapWord

	compatdataHeader := widget.NewRichTextFromMarkdown("### Proton Prefix Paths (compatdata)")
	compatdataInfoText := widget.NewLabel("Configure the Proton prefix directory (usually in steamapps/compatdata/[AppID]).")
	compatdataInfoText.Wrapping = fyne.TextWrapWord

	closeButton := widget.NewButton("Close", func() {
		settingsWindow.Close()
	})

	mainHeader := widget.NewRichTextFromMarkdown("## Settings")
	topInfo := widget.NewLabel("Configure custom paths for your games. Leave empty to use auto-discovery.")
	topInfo.Wrapping = fyne.TextWrapWord

	content := container.NewBorder(
		container.NewVBox(mainHeader, topInfo, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewHBox(layout.NewSpacer(), closeButton)),
		nil,
		nil,
		container.NewVScroll(
			container.NewVBox(
				gameHeader,
				gameInfoText,
				gameForm,
				widget.NewSeparator(),
				compatdataHeader,
				compatdataInfoText,
				compatdataForm,
			),
		),
	)

	settingsWindow.SetContent(content)
	return settingsWindow
}
