# Figma Discord Rich Presence

Current release target: **v2.0.0**

This is a Go app that updates your Discord status while you work in Figma Desktop.
Figma Desktop does not include built-in Discord Rich Presence support, so this project provides a plug-and-play option.

If you have suggestions, open an issue! :)

| Working in File  | Home  | Default Without Plugin |
| :---: | :---: | :---: |
| <img width="300" alt="Home" src="https://github.com/user-attachments/assets/1058ef1e-26b5-4176-8329-dfc1c5e2c22f" /> | <img width="300" alt="Working in File" src="https://github.com/user-attachments/assets/5c1c1402-303e-4180-8552-98bbdb45c8fe" /> | <img width="300" alt="Default" src="https://github.com/user-attachments/assets/2ace9e79-c094-483e-9cbd-55e765eea5e1" /> |

## Features

- **Project and file support**: Shows the specific project or file you are working on.
- **Browsing state**: Detects when you are in "Home" or working in a file.
- **Smart reconnect**: Automatically reconnects to Discord if the connection is lost or Discord restarts.
- **Cross-platform**: Native support for **Windows** and **macOS**.
- **Zero config**: Just run it (with Discord desktop running).

## Installation

### Website
1. Visit [Figma Discord Rich Presence](https://sleepypandas.github.io/Figma-Discord-Rich-Presence/).

### Windows
1. Go to the [Releases](https://github.com/SleepyPandas/Figma-Discord-Rich-Presence/releases) page.
2. Download `FigmaRPC-Windows-Installer.exe`.
3. Run the installer.
4. Optional: check "Run on Startup" during installation.

### macOS
1. Go to the [Releases](https://github.com/SleepyPandas/Figma-Discord-Rich-Presence/releases) page.
2. Download `FigmaRPC-MacOs-Installer.pkg`.
3. Run the installer (`Right Click -> Open` if you see a security warning).
4. The app installs and starts automatically.
5. If it does not start, open Terminal and run `figma-rpc`.

> [!IMPORTANT]
> On macOS, grant Accessibility permission so the app can read Figma window titles.
> Go to **System Settings -> Privacy & Security -> Accessibility** and add:
> **`/usr/local/bin/figma-rpc`**

> [!WARNING]
> **Windows and macOS may warn that the installer/app is unsigned or from an unidentified publisher.**
> This is expected because the app is currently unsigned.

<img width="500" height="255" alt="image" src="https://github.com/user-attachments/assets/34a93241-1594-4e9d-90ed-b58fe48135a5" />

## Usage

Once installed, the application runs in the background.
- Open **Figma**.
- Open **Discord**.
- Your status should update within a few seconds.

## Development

If you want to build this yourself:

### Prerequisites
- [Go 1.25+](https://go.dev/dl/)
- **Windows**: [Inno Setup 6](https://jrsoftware.org/isdl.php) (for building the installer)
- **macOS**: Xcode Command Line Tools

### Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/SleepyPandas/Figma-Discord-Rich-Presence.git
   cd Figma-Discord-Rich-Presence
   ```
2. No `.env` file is required for the current build.

### Build
**Windows**:
```bash
cd src
go build -ldflags "-H windowsgui" -o ../figma-rpc.exe .
```

**macOS**:
```bash
cd src
go build -o ../figma-rpc .
```

## License

This project is licensed under the Apache 2.0 License. See [LICENSE](LICENSE).
