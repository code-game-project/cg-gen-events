# cg-gen-events
![CGE Version](https://img.shields.io/badge/CGE-v0.4-green)

Easily generate [CodeGame](https://code-game.org) event definitions for a variety of languages from [CodeGame Events (CGE)](https://docs.code-game.org/specifications/cge) files.

## Installation

cg-gen-events should be used through [codegame-cli](https://github.com/code-game-project/codegame-cli):
```
codegame gen-events ...
```

## Usage

Generate from a local file:
```sh
codegame gen-events my_game.cge
```

Generate from a remote file:
```sh
codegame gen-events https://example.com
# same as
codegame gen-events https://example.com/events
```

Specify an output directory:
```sh
codegame gen-events -o events/ my_game.cge
```

Specify a list of languages as command line arguments instead of the interactive menu:
```sh
codegame gen-events -l go,typescript my_game.cge
```

Use `codegame gen-events --help` for a complete list of available options.

## Supported languages

- C#
- Go
- Java
- Markdown docs
- TypeScript
- JSON

## Editor Support

- LSP: [cge-ls](https://github.com/code-game-project/cg-gen-events/blob/main/cmd/cge-ls/README.md)
- VS Code: [vscode-codegame](https://github.com/code-game-project/vscode-codegame)
- Vim: [vim-codegame](https://github.com/code-game-project/vim-codegame)

## Building

### Prerequisites

- [Go](https://go.dev/) 1.18+

```sh
git clone https://github.com/code-game-project/cg-gen-events
cd cg-gen-events
go build -o cg-gen-events ./cmd/cg-gen-events
```

## License

Copyright (c) 2022-2023 Julian Hofmann

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
