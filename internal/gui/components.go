package gui

import (
	"fmt"
	"image/color"
	"net/url"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	"github.com/bazsalanszky/fusioncore/internal/vfs"
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
			
			activateBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil)
			activateBtn.Importance = widget.HighImportance
			deactivateBtn := widget.NewButtonWithIcon("", theme.MediaPauseIcon(), nil)
			uninstallBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			uninstallBtn.Importance = widget.DangerImportance
			
			headerRow := container.NewHBox(
				statusIndicator,
				modName,
				layout.NewSpacer(),
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
			activateBtn := headerRow.Objects[3].(*widget.Button)
			deactivateBtn := headerRow.Objects[4].(*widget.Button)
			uninstallBtn := headerRow.Objects[5].(*widget.Button)
			
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

			activateBtn.OnTapped = func() {
				vfs.Activate(m.Name)
				state.mods, _ = mod.LoadMods()
				modList.Refresh()
			}
			deactivateBtn.OnTapped = func() {
				vfs.Deactivate(m.Name)
				state.mods, _ = mod.LoadMods()
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
					mod.SaveMods(newMods)
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
	
	launchButton := widget.NewButtonWithIcon("Launch Fallout 76", theme.MediaPlayIcon(), nil)
	launchButton.Importance = widget.HighImportance
	
	return usernameLabel, launchButton
}
