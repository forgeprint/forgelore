# forgelore

Local-first, team-shared memory for coding agents. It remembers the fix for
an error and injects it only when the same error comes back.

```sh
npm install -g forgelore
forgelore init
```

This package is a wrapper. The program is a single static Go binary with no
runtime dependencies; npm installs the one built for your platform as an
ordinary optional dependency, and the wrapper runs it. There is no
postinstall script and nothing is downloaded at install time, so
`--ignore-scripts` works.

Installing without npm works too, and is the same program:

```sh
curl -fsSL https://raw.githubusercontent.com/forgeprint/forgelore/main/scripts/install.sh | bash
```

Everything else — what it does, how it measures whether it paid for itself,
and how to wire it to an agent — is in the
[repository](https://github.com/forgeprint/forgelore).
