<!-- status: reviewed -->
<!-- search: true -->

# Install

```sh
curl -fsSL https://ant-lang.org/install.sh | sh
```

The script downloads `ant` into `/usr/local/bin`, or `~/bin` under Git Bash.

On Windows, in PowerShell:

```powershell
irm https://ant-lang.org/install.ps1 | iex
```

The script downloads `ant.exe` into `~\bin` and adds that directory to your `PATH`.

Or build it from source:

```sh
go install ant-lang.org/cmd/ant@latest
```

## Run a program

```sh
ant check main.ant
ant run main.ant
```

Without a file, `ant` opens the REPL. `ant fmt` formats sources, `ant test` runs the test files
beside them, and `ant doc` prints the documentation of a name:

```sh
ant fmt .
ant test
ant doc Array.map
```

## Shell completion

Complete commands, flags, and files in the shell:

```sh
source <(ant --comp zsh)   # in ~/.zshrc, after compinit
source <(ant --comp bash)  # in ~/.bashrc
ant --comp fish | source   # in ~/.config/fish/config.fish
```

Open a new shell, or source its startup file, for the line to take effect. zsh ships a completion
named `_ant` for Apache Ant; source Ant's script after `compinit` so it replaces that one.

## Next steps

- [Your first program](first-program.md)
- [ant check](../tools/check.md)
- [ant run and the REPL](../tools/run.md)
