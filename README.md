# cg-gen-events
![CGE Version](https://img.shields.io/badge/CGE-v0.2-green)

Easily generate [CodeGame](https://github.com/code-game-project) event definitions for a variety of languages from [CodeGame Events (CGE)](https://github.com/code-game-project/docs/blob/main/docs/code-game-events-language-specification.md) files.

## Installation

### Prebuilt binaries

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
cg-gen-events -output events/ my_game.cge
```

Specify a list of languages as command line arguments instead of the interactive menu:
```sh
cg-gen-events -languages go,typescript my_game.cge
```

Use `-help` for a complete list of available options.

## Supported languages

- Go
- Markdown docs
- TypeScript

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
