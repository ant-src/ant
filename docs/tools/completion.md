<!-- status: reviewed -->

# Shell completion

```sh
ant --comp bash|zsh|fish
```

`ant --comp` prints a script that wires the shell's Tab completion to `ant`. Load the script once in
the shell's startup file, and Tab completes `ant`'s commands, flags, and files.

## Set up

Add one line to the startup file of your shell:

```sh
source <(ant --comp bash)  # in ~/.bashrc
source <(ant --comp zsh)   # in ~/.zshrc, after compinit
ant --comp fish | source   # in ~/.config/fish/config.fish
```

Open a new shell, or source the file, for the line to take effect.

zsh ships a completion named `_ant` for Apache Ant. Source the script after `compinit`, as above, so
it replaces that one. To load it from `$fpath` instead, save it as `_ant` in a directory that comes
before the system's:

```sh
mkdir -p ~/.zfunc
ant --comp zsh > ~/.zfunc/_ant
# in ~/.zshrc, before compinit:
fpath=(~/.zfunc $fpath)
```

## What it completes

The script asks `ant` itself for the candidates each time you press Tab, so a newer `ant` completes
its new commands and flags without a new script.

| You type        | Tab offers                                              |
| --------------- | ------------------------------------------------------- |
| `ant `          | the commands, `.ant` files, and directories             |
| `ant -`         | `-h`, `--help`, `--version`, `--comp`                   |
| `ant --comp `   | `bash`, `zsh`, `fish`                                   |
| `ant run `      | `.ant` files, then any file for the program's arguments |
| `ant main.ant ` | any file, for the program's arguments                   |
| `ant check `    | `.ant` files                                            |
| `ant fmt `      | `.ant` files, for each path                             |
| `ant fmt -`     | `--check`                                               |
| `ant test `     | `_test.ant` files, for each path                        |
| `ant test -`    | the flags of `ant test`, such as `-run` and `-v`        |
| `ant doc `      | `.ant` files                                            |

Wherever files are offered, directories are too, so you can complete a path one directory at a time:

```sh
$ ant test src/<Tab>
src/geo_test.ant
```

## Details

- A file or directory whose name starts with a dot is offered only when the word you type starts
  with a dot.
- A path that starts with `~/` completes from your home directory.
- `ant doc` completes the file argument only, not the names of the standard library.
- `ant --comp` with no shell, or with a shell other than `bash`, `zsh`, or `fish`, prints an error
  and exits with status 2:

  ```sh
  $ ant --comp pwsh
  unknown shell "pwsh"; want bash, zsh, or fish
  ```

## See also

- [Install](../start/install.md)
- [ant run and the REPL](run.md)
- [ant test](test.md)
- [ant fmt](fmt.md)
- [ant doc](doc.md)
