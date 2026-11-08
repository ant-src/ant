<!-- status: draft -->

# Editor setup

The language server and the debugger come with the binary, and a plugin brings them to the editor.
Each plugin finds `ant` on `PATH` or in your Go bin directory.

## VS Code

Install
[Ant from the Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=ant-lang.ant-lang),
or download `ant.vsix` from a [release](https://github.com/ant-src/vscode-ant/releases) and run
**Extensions: Install from VSIX…** from the Command Palette. `ant.path` in the settings names
another `ant`. [What the extension does](/vscode).

## IntelliJ IDEA

Install
[Ant from the JetBrains Marketplace](https://plugins.jetbrains.com/plugin/35684-ant-language), or
search for Ant under **Settings | Plugins | Marketplace**. It needs IntelliJ IDEA 2025.3 or later,
and 2026.1 for rename. **Settings | Languages & Frameworks | Ant** names another `ant`.
[What the plugin does](/intellij).

## Neovim

With [lazy.nvim](https://github.com/folke/lazy.nvim):

```lua
{ "ant-src/nvim-ant" }
```

With Neovim's packages:

```sh
git clone https://github.com/ant-src/nvim-ant ~/.local/share/nvim/site/pack/ant/start/nvim-ant
```

It needs Neovim 0.11 or later, and [nvim-dap](https://github.com/mfussenegger/nvim-dap) for
debugging. The plugin needs no `setup()`, and `:checkhealth ant` says what it found.
`vim.g.ant = { path = "..." }` names another `ant`. [What the plugin does](/nvim).
