# cg-gen-events
![CGE Version](https://img.shields.io/badge/CGE-v0.3-green)

Easily generate [CodeGame](https://github.com/code-game-project) event definitions for a variety of languages from [CodeGame Events (CGE)](https://github.com/code-game-project/docs/blob/main/docs/code-game-events-language-specification.md) files.

## Usage

Generate from a local file:
```sh
cg-gen-events my_game.cge
```

Generate from a remote file:
```sh
cg-gen-events https://example.com
# same as
cg-gen-events https://example.com/events
```

Specify an output directory:
```sh
cg-gen-events --output events/ my_game.cge
```

Specify a list of languages as command line arguments instead of the interactive menu:
```sh
cg-gen-events --languages go,typescript my_game.cge
```

Use `--help` for a complete list of available options.

## Supported languages

- Go
- Markdown docs
- TypeScript

## Installation

### Windows

1. Open the Start menu
2. Search for `powershell`
3. Hit `Run as Administrator`
4. Paste the following commands and hit enter:

#### Install

```powershell
Invoke-WebRequest -Uri "https://github.com/code-game-project/cg-gen-events/releases/latest/download/cg-gen-events-windows-amd64.zip" -OutFile "C:\Program Files\cg-gen-events.zip"
Expand-Archive -LiteralPath "C:\Program Files\cg-gen-events.zip" -DestinationPath "C:\Program Files\cg-gen-events"
rm "C:\Program Files\cg-gen-events.zip"
Set-ItemProperty -Path 'Registry::HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment' -Name PATH -Value "$((Get-ItemProperty -Path 'Registry::HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\Environment' -Name PATH).path);C:\Program Files\cg-gen-events"
```

**IMPORTANT:** Please reboot for the installation to take effect.

#### Update

```powershell
rm -r -fo "C:\Program Files\cg-gen-events"
Invoke-WebRequest -Uri "https://github.com/code-game-project/cg-gen-events/releases/latest/download/cg-gen-events-windows-amd64.zip" -OutFile "C:\Program Files\cg-gen-events.zip"
Expand-Archive -LiteralPath "C:\Program Files\cg-gen-events.zip" -DestinationPath "C:\Program Files\cg-gen-events"
rm "C:\Program Files\cg-gen-events.zip"
```

### macOS

Open the Terminal application, paste the command for your architecture and hit enter.

To update, simply run the command again.

#### x86_64

```sh
curl -L https://github.com/code-game-project/cg-gen-events/releases/latest/download/cg-gen-events-darwin-amd64.tar.gz | tar -xz cg-gen-events && sudo mv cg-gen-events /usr/local/bin
```

#### ARM64

```sh
curl -L https://github.com/code-game-project/cg-gen-events/releases/latest/download/cg-gen-events-darwin-arm64.tar.gz | tar -xz cg-gen-events && sudo mv cg-gen-events /usr/local/bin
```

### Linux

Open a terminal, paste the command for your architecture and hit enter.

To update, simply run the command again.

#### x86_64

```sh
curl -L https://github.com/code-game-project/cg-gen-events/releases/latest/download/cg-gen-events-linux-amd64.tar.gz | tar -xz cg-gen-events && sudo mv cg-gen-events /usr/local/bin
```

#### ARM64

```sh
curl -L https://github.com/code-game-project/cg-gen-events/releases/latest/download/cg-gen-events-linux-arm64.tar.gz | tar -xz cg-gen-events && sudo mv cg-gen-events /usr/local/bin
```

### Other

You can download a prebuilt binary file for your operating system on the [releases](https://github.com/code-game-project/cg-gen-events/releases) page.

You might need to make the file executable before running it.

### Compiling from source

#### Prerequisites

- [Go](https://go.dev/) 1.18+

```sh
git clone https://github.com/code-game-project/cg-gen-events.git
cd cg-gen-events
go build .
```

## License

Copyright (c) 2022 CodeGame Contributors (https://github.com/orgs/code-game-project/people)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
