#!/usr/bin/env bash
# Assembles the npm packages from what release.sh built: one package per
# platform carrying the binary, and one wrapper that depends on all of them
# and runs whichever npm installed (ADR-0024).
#
# Like release.sh, it does not publish. It prints the commands.
set -euo pipefail
cd "$(dirname "$0")/.."

version="${1:-}"
if [ -z "$version" ]; then
	echo "usage: $0 <version>   e.g. $0 v0.1.4" >&2
	exit 1
fi
case "$version" in
v[0-9]*.[0-9]*.[0-9]*) ;;
*)
	echo "a version is vMAJOR.MINOR.PATCH, such as v0.1.4; got $version" >&2
	exit 1
	;;
esac
# npm versions have no leading v, and npm is strict about it.
npm_version="${version#v}"

if [ ! -f dist/SHA256SUMS ]; then
	echo "dist/ holds no release build; run ./scripts/release.sh $version first" >&2
	exit 1
fi

# dist/ is not trusted to still be what release.sh built. Anything that runs
# crosscheck.sh afterwards — ci.sh does — rewrites the binaries in place with
# a build stamped from `git describe`, and SHA256SUMS is left behind
# describing bytes that are gone. Packing those publishes a dev build under a
# release version, and it happened: forgelore@0.1.4 reached npm carrying
# v0.1.4-1-g9b2beba-dirty.
#
# The checksum file catches every case, including the cross-built targets
# that cannot be run here.
echo "==> checking dist/ is still the $version build"
if command -v sha256sum >/dev/null 2>&1; then
	checksums_ok() { (cd dist && sha256sum --check --quiet SHA256SUMS); }
else
	checksums_ok() { (cd dist && shasum -a 256 --check --status SHA256SUMS); }
fi
if ! checksums_ok; then
	cat >&2 <<EOF
dist/ no longer matches its own SHA256SUMS: something rebuilt the binaries
after release.sh wrote it. Build the release again before packing:

  ./scripts/release.sh $version
EOF
	exit 1
fi

# And the version is read back out of a binary rather than assumed, because
# matching checksums only prove dist/ is internally consistent. The host's
# own target is the one that can be run.
case "$(uname -s)/$(uname -m)" in
Darwin/arm64) host=dist/forgelore_darwin_arm64 ;;
Darwin/x86_64) host=dist/forgelore_darwin_amd64 ;;
Linux/aarch64 | Linux/arm64) host=dist/forgelore_linux_arm64 ;;
Linux/x86_64) host=dist/forgelore_linux_amd64 ;;
*) host="" ;;
esac
if [ -n "$host" ]; then
	stamped=$("$host" version | awk '{print $2}')
	if [ "$stamped" != "$version" ]; then
		echo "$host says $stamped, not $version; rebuild with ./scripts/release.sh $version" >&2
		exit 1
	fi
fi

# npm names an architecture the way Node reports it, not the way Go does.
# The mapping is here and nowhere else, because the wrapper reads
# process.platform and process.arch straight out of Node.
targets="linux_amd64:linux:x64 linux_arm64:linux:arm64 darwin_amd64:darwin:x64 darwin_arm64:darwin:arm64 windows_amd64:win32:x64 windows_arm64:win32:arm64"

out=dist/npm
rm -rf "$out"
mkdir -p "$out"

optional=""
for t in $targets; do
	goname="${t%%:*}"
	rest="${t#*:}"
	os="${rest%%:*}"
	cpu="${rest#*:}"

	src="dist/forgelore_$goname"
	exe="forgelore"
	if [ "$os" = "win32" ]; then
		src="$src.exe"
		exe="forgelore.exe"
	fi
	[ -f "$src" ] || { echo "missing $src" >&2; exit 1; }

	pkg="forgelore-$os-$cpu"
	mkdir -p "$out/$pkg"
	# The executable bit is part of the tarball. Without it npm installs a
	# file nobody can run, and the wrapper's error would blame the wrong
	# thing.
	install -m 755 "$src" "$out/$pkg/$exe"

	# No "exports": the wrapper resolves the binary by subpath, and an
	# exports map would have to list it. No "files" either, because the
	# package is two files and both belong.
	cat > "$out/$pkg/package.json" <<EOF
{
  "name": "$pkg",
  "version": "$npm_version",
  "description": "The forgelore binary for $os $cpu. Installed by the forgelore package; not useful on its own.",
  "repository": { "type": "git", "url": "git+https://github.com/forgeprint/forgelore.git" },
  "license": "Apache-2.0",
  "os": ["$os"],
  "cpu": ["$cpu"],
  "preferUnplugged": true
}
EOF
	optional="$optional    \"$pkg\": \"$npm_version\",
"
done

# Trailing comma off the last dependency.
optional="${optional%,
}"

mkdir -p "$out/forgelore"
cp npm/forgelore.js "$out/forgelore/forgelore.js"
chmod 755 "$out/forgelore/forgelore.js"
cp npm/README.md "$out/forgelore/README.md"
cp LICENSE "$out/forgelore/LICENSE"
cat > "$out/forgelore/package.json" <<EOF
{
  "name": "forgelore",
  "version": "$npm_version",
  "description": "Local-first, team-shared memory for coding agents: remembers a fix and injects it when the same error comes back.",
  "keywords": ["memory", "errors", "tokens", "claude-code", "agent"],
  "homepage": "https://github.com/forgeprint/forgelore#readme",
  "repository": { "type": "git", "url": "git+https://github.com/forgeprint/forgelore.git" },
  "license": "Apache-2.0",
  "bin": { "forgelore": "forgelore.js" },
  "optionalDependencies": {
$optional
  }
}
EOF

# The tarballs stay inside dist/npm. release.sh tells whoever publishes by
# hand to upload dist/*, and a .tgz caught by that glob would ship npm
# packaging as a release asset.
echo "==> packing"
for d in "$out"/*/; do
	(cd "$d" && npm pack --silent --pack-destination ..) > /dev/null
done

echo
echo "dist/npm holds $(find "$out" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ') packages:"
for d in "$out"/*/; do
	echo "  $(basename "$d")"
done
echo

# Each tarball is named, rather than globbed. dist/npm/forgelore-* matches
# the wrapper's own tarball and every platform tarball as well as the
# directories, so a glob here would either publish the wrapper first — the
# one order that leaves a user with no binary — or try to cd into a .tgz.
publish=""
for t in $targets; do
	rest="${t#*:}"
	publish="$publish  npm publish --access public dist/npm/forgelore-${rest%%:*}-${rest#*:}-$npm_version.tgz
"
done

cat <<EOF
Publishing is a human action, as it is for a release. The platform packages
go first: the wrapper depends on them by exact version, and a wrapper on the
registry whose dependencies are not there yet installs with no binary.

$publish
  npm publish --access public dist/npm/forgelore-$npm_version.tgz

Then check what a user gets, in an empty directory:

  npm install forgelore@$npm_version && ./node_modules/.bin/forgelore version
EOF
