<!-- status: draft -->

# Install

```sh
curl https://ant-lang.org/install.sh | sh
```

The script downloads `ant` into `/usr/local/bin`.

Or build it from source:

```sh
go install github.com/ant-src/ant/cmd/ant@latest
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
ant test .
ant doc Array.map
```

## Shell completion

Complete commands, flags, and files in the shell:

```sh
source <(ant --comp zsh)   # in ~/.zshrc, after compinit
source <(ant --comp bash)  # in ~/.bashrc
ant --comp fish | source   # in ~/.config/fish/config.fish
```
