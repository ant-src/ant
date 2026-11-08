<!-- status: draft -->

# Install

```sh
curl -fsSL https://ant-lang.org/install.sh | sh
```

The script downloads the latest release's `ant` for your OS and architecture into `/usr/local/bin`,
or `~/bin` in Git Bash on Windows. `ANT_VERSION=vX.Y.Z` installs that release, and `ANT_INSTALL_DIR`
picks another directory.

With Go installed, `go install` builds it from source instead:

```sh
go install github.com/ant-src/ant/cmd/ant@latest
```

`ant version` prints the release you have.

## Run a program

`ant check` type-checks a file with its imports, and `ant run` runs it:

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
