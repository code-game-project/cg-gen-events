# cge-ls
![CGE Version](https://img.shields.io/badge/CGE-v0.5-green)

An [LSP](https://microsoft.github.io/language-server-protocol) implementation for [CGE](https://github.com/Bananenpro/cg-gen-events) files.

## Features

- [x] diagnostics
- [x] code completion
- [ ] snippets
- [ ] goto definition
- [ ] symbol rename

## Installation

cge-ls should be used through [codegame-cli](https://github.com/code-game-project/codegame-cli):
```
codegame lsp cge
```

### VS Code

Install the [vscode-codegame](https://github.com/code-game-project/vscode-codegame#installation) extension.

### Neovim

Install the [vim-codegame](https://github.com/code-game-project/vim-codegame#installation) plugin for syntax highlighting and indentation.

#### coc

In [`coc-settings.json`](https://github.com/neoclide/coc.nvim/wiki/Language-servers#register-custom-language-servers):
```json
{
  "languageserver": {
    "cge-ls": {
      "command": "cge-ls",
      "filetypes": ["cge"],
      "rootPatterns": [".git/", "."]
    }
  }
}
```

#### lspconfig

In `init.lua`:
```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')
configs.cge = {
  default_config = {
    cmd = { "cge-ls" },
    root_dir = lspconfig.util.root_pattern('.git'),
    single_file_support = true,
    filetypes = { 'cge' },
    init_options = {
      command = { 'cge-ls' },
    },
  },
}
lspconfig.cge.setup{}
```

## Building

### Prerequisites

- [Go](https://go.dev) 1.18+

```
git clone https://github.com/code-game-project/cg-gen-events
cd cg-gen-events
go build -o cge-ls ./cmd/cge-ls
```

## License

Copyright (c) 2022 Julian Hofmann

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
