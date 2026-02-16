# Figma Discord Rich Presence

This is a simple tool written in Go to update your discord status. Figma Desktop has no built in tool or plugin to change your status. And the existing tools were not so plug and play so I thought I would make one. 
Also it annoyed me that someone I know didn't have RPC on Figma but they use it everyday... 

If you have any suggestions make an issue! :) 

| Home | Working in File | Default Without Plugin |
| :---: | :---: | :---: |
| <img width="300" alt="Home" src="https://github.com/user-attachments/assets/1058ef1e-26b5-4176-8329-dfc1c5e2c22f" /> | <img width="300" alt="Working in File" src="https://github.com/user-attachments/assets/5c1c1402-303e-4180-8552-98bbdb45c8fe" /> | <img width="300" alt="Default" src="https://github.com/user-attachments/assets/2ace9e79-c094-483e-9cbd-55e765eea5e1" /> |


## Features

- **Project & File Support**: Shows the specific project or file you are working on.
- **Browsing State**: Detects when you are in the "Home" or "Working on a Project".
- **Smart Reconnect**: Automatically reconnects to Discord if the connection is lost or if you restart Discord.
- **Cross-Platform**: Native support for **Windows** and **macOS**.
- **Zero Config**: Just run it, and it works (assuming you have the Discord desktop app running).

## Installation 

### Website 
1. Visit the github pages website for distribution (Coming Soon!)

### Windows
1. Go to the [Releases](https://github.com/SleepyPandas/Figma-Discord-Rich-Presence/releases) page.
2. Download the latest `FigmaRPC_Setup.exe`.
3. Run the installer.
4. (Optional) Check "Run on Startup" during installation to have it alwys running.

### macOS
1. Go to the [Releases](https://github.com/SleepyPandas/Figma-Discord-Rich-Presence/releases) page.
2. Download the latest `FigmaRPC_Installer.pkg`.
3. Run the installer (`Right Click -> Open` if you see a security warning).
4. The app will install and start automatically.


> [!NOTE]
> To uninstall on Windows go into Add or Remove Programs from the start Menu and Search for Figma Discord Rich Prescence
> 
> MacOs will warn you about unverified publisher this is normal as it would cost me 99 USD$ to verify it...

<img width="500" height="255" alt="image" src="https://github.com/user-attachments/assets/34a93241-1594-4e9d-90ed-b58fe48135a5" />

## Usage

Once installed, the application runs in the background. 
- Open **Figma**.
- Open **Discord**.
- Your status should update within a few seconds!

## Development

If you want to build this yourself:

### Prerequisites
- [Go 1.21+](https://go.dev/dl/)
- **Windows**: [Inno Setup 6](https://jrsoftware.org/isdl.php) (for building the installer)
- **macOS**: Xcode Command Line Tools

### Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/SleepyPandas/Figma-Discord-Rich-Presence.git
   cd Figma-Discord-Rich-Presence
   ```
2. Create a `.env` file in `src/` with your Discord Client ID:
   _Discord_clientID is already public id but I used an env because its cool you could optinally just skip this and hard code it in the main.go_
   ```env
   DISCORD_CLIENT_ID=your_client_id_here
   ```

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

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
