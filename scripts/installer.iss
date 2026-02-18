; Figma Discord Rich Presence - Inno Setup Installer Script
; Produces a professional Windows installer with GUI wizard

[Setup]
AppName=Figma Discord Rich Presence
AppVersion=1.0.1
AppPublisher={#GetEnv('APP_PUBLISHER')}
AppPublisherURL=https://github.com/SleepyPandas/Figma-Discord-Rich-Presence
DefaultDirName={autopf}\Figma Discord RPC
DefaultGroupName=Figma Discord RPC
OutputDir=..\dist
OutputBaseFilename=FigmaRPC_Setup
Compression=lzma
SolidCompression=yes
UninstallDisplayName=Figma Discord Rich Presence
PrivilegesRequired=lowest
WizardStyle=modern

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "startuplaunch"; Description: "Run on Windows startup"; GroupDescription: "Additional options:"
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; The main executable (built with -H windowsgui)
Source: "..\figma-rpc.exe"; DestDir: "{app}"; Flags: ignoreversion
; The .env config file
Source: "..\src\.env"; DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist

[Icons]
; Start Menu shortcut
Name: "{group}\Figma Discord RPC"; Filename: "{app}\figma-rpc.exe"
Name: "{group}\Uninstall Figma Discord RPC"; Filename: "{uninstallexe}"
; Desktop shortcut (optional)
Name: "{autodesktop}\Figma Discord RPC"; Filename: "{app}\figma-rpc.exe"; Tasks: desktopicon

[Registry]
; Run on startup (only if user checks the box)
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "FigmaDiscordRPC"; ValueData: """{app}\figma-rpc.exe"""; Flags: uninsdeletevalue; Tasks: startuplaunch

[Run]
; Offer to launch the app after install
Filename: "{app}\figma-rpc.exe"; Description: "Launch Figma Discord Rich Presence"; Flags: nowait postinstall skipifsilent
