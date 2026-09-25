# Contributing to Forgelore

## Sign your commits off (DCO)

This project uses the [Developer Certificate of Origin](DCO). Every commit must
carry a `Signed-off-by` line matching the author:

```sh
git commit -s -m "your message"
```

There is no CLA. The sign-off is the whole agreement.

## Dependencies

Forgelore depends on the Go standard library and `modernc.org/sqlite`, and
nothing else. This is deliberate: the binary has to build and run on a machine
with no toolchain, no C compiler and no network.

If a change appears to need another dependency, open an issue first. If it is
accepted, it comes with an ADR in `docs/adr/` explaining why, and the module is
vendored into `vendor/`.

## Before you open a pull request

Run the same checks CI runs:

```sh
./scripts/ci.sh
```

That is formatting, `go vet`, tests, a cross-compile of every release target
with `CGO_ENABLED=0`, and a secret scan. CI calls this exact script, so a green
run locally means a green run there.

## Agent adapters are data, not code

Support for a coding agent lives in a mapping file under `adapters/`, embedded
into the binary and overridable by a file in the user's project. Adding or
fixing support for an agent should not require writing Go, and changing an
event name should not require a new release.

An adapter change needs real event samples in `testdata/agents/<agent>/<version>/`
and a contract test that runs against them. Claims about an agent's hook format
must cite that agent's current official documentation.

## Do not commit secrets or captured data

`scripts/gitleaks.sh` runs on every pull request. Event samples and error
samples under `testdata/` must be scrubbed of paths, hostnames, tokens and
anything else identifying before they are committed.
