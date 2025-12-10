
# Fusion Core ☢️

### Powering Fallout Mods on Linux

**Fusion Core** is a native Linux mod manager designed specifically for Bethesda games (Fallout 76, Fallout 4, New Vegas) running under Valve's Proton.

Unlike other solutions that require complex Wine wrappers or running Windows mod managers inside the same bottle as the game, Fusion Core runs natively on your Linux OS, managing files and configurations from the outside in.

-----

## 🚀 The Problem

Modding Fallout on Linux is traditionally a headache. You usually have to:

1.  Install a Windows Mod Manager (Vortex/MO2) inside the specific Proton prefix of the game.
2.  Deal with broken UI rendering or .NET dependencies in Wine.
3.  Struggle to handle `nxm://` links from your Linux browser to the Wine application.

## 💡 The Solution

**Fusion Core** simplifies this by treating the Proton Prefix (`compatdata`) as just another directory.

  * **Native Performance:** Written in **Go**, it runs lightweight and fast on your actual OS.
  * **Symlink Magic:** It downloads mods to a native Linux directory and "injects" them into the game using Symlinks. Wine handles Linux symlinks transparently, so the game sees the files, but your game folder stays clean.
  * **Protocol Handling:** Captures "Download with Manager" links from Nexus Mods directly in your Linux browser.

-----

## 🛠️ Tech Stack

  * **Language:** Go (Golang)
  * **GUI:**  [Fyne](https://fyne.io/)

-----

## 🗺️ Roadmap

### Phase 1: Core Systems (Current Focus)

  - [x] **Prefix Detection:** Auto-detect Steam Library and locate `compatdata` for Fallout 76 (AppID: `1151340`).
  - [x] **Nexus API:** Register `nxm://` protocol handler on Linux (KDE/Gnome compatible) and parse download tokens.
  - [x] **VFS (Virtual File System):** Implement the Symlink logic to map a central "Mods" folder into the Proton `Data` folder.

### Phase 2: Configuration Management

  - [x] **INI Parser:** Automatically generate/update `Fallout76Custom.ini` to register `.ba2` archives.
  - [ ] **Load Order:** Basic UI to drag-and-drop load order (updates `plugins.txt`).

### Phase 3: GUI & Polish

  - [ ] **Mod Library UI:** Visual grid of installed mods with metadata (Version, Author, Description).
  - [ ] **Profiles:** Switch between "Adventure Mode" and "Nuclear Winter" (RIP) or other mod setups instantly.

-----

## ⚡ Getting Started (For Developers)

### Prerequisites

  * Go 1.21+
  * Steam (installed via native package or Flatpak\*)
  * *Note: Flatpak support requires permission tuning to access the filesystem.*

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/fusion-core.git

# Enter directory
cd fusion-core

# Build the binary
go build -o fusion-core cmd/main.go

# Run
./fusion-core
```

-----

## 📂 Directory Structure Strategy

Fusion Core keeps your install clean.

  * **Mod Storage:** `~/Games/FusionCore/Mods/Fallout76/` (Where the actual files live)
  * **Game Folder:** `.../steamapps/common/Fallout76/Data/` (Where we place Symlinks)

-----

## 🤝 Contributing

Pull requests are welcome\! If you know Go, Linux file systems, or the Nexus API, please help us build the best modding experience for Linux gamers.

## 📜 License

MIT License.

-----

*Ad Victoriam

-----
