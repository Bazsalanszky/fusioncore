package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/bazsalanszky/fusioncore/internal/mod"
	"github.com/bazsalanszky/fusioncore/internal/vfs"
	"net/url"
	"os"
)

func newAPIKeyWindow(a fyne.App, onSave func(string)) fyne.Window {
	w := a.NewWindow("Nexus API Key")
	w.Resize(fyne.NewSize(400, 200))

	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Enter your API key")

	infoLabel := widget.NewLabel("You can get your API key from the Nexus Mods website.")

	nexusURL, _ := url.Parse("https://www.nexusmods.com/users/myaccount?tab=api")
	hyperlink := widget.NewHyperlink("Get API Key", nexusURL)

	saveButton := widget.NewButton("Save", func() {
		onSave(apiKeyEntry.Text)
		w.Close()
	})

	w.SetContent(container.NewVBox(
		infoLabel,
		hyperlink,
		apiKeyEntry,
		saveButton,
	))

	return w
}

func showErrorDialog(err error, w fyne.Window) {
	dialog.ShowError(err, w)
}

func newModList(w fyne.Window, state *AppState) (*widget.List, []*mod.Mod) {
	var modList *widget.List
	modList = widget.NewList(
		func() int {
			return len(state.mods)
		},
		func() fyne.CanvasObject {
			activateButton := widget.NewButton("Activate", nil)
			deactivateButton := widget.NewButton("Deactivate", nil)
			uninstallButton := widget.NewButton("Uninstall", nil)
			return container.NewBorder(
				nil,
				nil,
				nil,
				container.NewHBox(
					activateButton,
					deactivateButton,
					uninstallButton,
				),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			m := state.mods[i]
			status := "inactive"
			if m.Active {
				status = "active"
			}
			c := o.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			hbox := c.Objects[1].(*fyne.Container)
			label.SetText(m.Name + " (" + status + ")")

			activateButton := hbox.Objects[0].(*widget.Button)
			deactivateButton := hbox.Objects[1].(*widget.Button)
			uninstallButton := hbox.Objects[2].(*widget.Button)

			activateButton.OnTapped = func() {
				vfs.Activate(m.Name)
				state.mods, _ = mod.LoadMods()
				modList.Refresh()
			}
			deactivateButton.OnTapped = func() {
				vfs.Deactivate(m.Name)
				state.mods, _ = mod.LoadMods()
				modList.Refresh()
			}
			uninstallButton.OnTapped = func() {
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
				activateButton.Disable()
				deactivateButton.Enable()
			} else {
				activateButton.Enable()
				deactivateButton.Disable()
			}
		},
	)

	return modList, state.mods
}

func newTopBar(w fyne.Window) (fyne.CanvasObject, *widget.Label, *widget.Button) {
	usernameLabel := widget.NewLabel("Fetching username...")
	launchButton := widget.NewButton("Launch Game", nil)

	top := container.NewVBox(
		container.NewHBox(usernameLabel, launchButton),
		widget.NewSeparator(),
	)

	return top, usernameLabel, launchButton
}
