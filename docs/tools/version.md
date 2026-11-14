<!-- status: reviewed -->

# ant version

`ant version` prints the version of the `ant` command: the release it was built from.

```sh
ant version
```

```text
ant v0.1.0
```

## Uses

- Check which `ant` your shell runs after you [install](../start/install.md) or update it.
- Include the line in a bug report, so that the report names the release that showed the bug.
- Print it in CI before the other steps, so that the log records the release that ran the checks and
  tests.

## See also

- [Install](../start/install.md)
- [ant run and the REPL](run.md)
- [ant check](check.md)
