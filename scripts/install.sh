#!/usr/bin/env bash
# Installs forgelore into ~/.local/bin, verifying the checksum first.
#
# No sudo, no package manager, nothing written outside your home directory.
# The binary is static and has no runtime dependencies, so there is nothing
# else to install alongside it.
#
#   curl -fsSL https://raw.githubusercontent.com/forgeprint/forgelore/main/scripts/install.sh | bash
#
# FORGELORE_VERSION picks a version other than the latest, and
# FORGELORE_BASE_URL points somewhere other than GitHub, which is how the
# acceptance test runs it against a local server.
set -euo pipefail

repo="forgeprint/forgelore"
install_dir="${FORGELORE_INSTALL_DIR:-$HOME/.local/bin}"
version="${FORGELORE_VERSION:-}"

fail() { echo "forgelore install: $*" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || fail "curl is required"

case "$(uname -s)" in
Linux) goos=linux ;;
Darwin) goos=darwin ;;
*) fail "unsupported operating system $(uname -s); build from source instead" ;;
esac
case "$(uname -m)" in
x86_64 | amd64) goarch=amd64 ;;
arm64 | aarch64) goarch=arm64 ;;
*) fail "unsupported architecture $(uname -m); build from source instead" ;;
esac

if [ -z "${FORGELORE_BASE_URL:-}" ]; then
	if [ -z "$version" ]; then
		# The redirect from /releases/latest names the tag, which avoids
		# parsing the API and needing a token for it.
		version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
			"https://github.com/$repo/releases/latest" | sed 's|.*/tag/||')"
		[ -n "$version" ] || fail "could not work out the latest version"
	fi
	base="https://github.com/$repo/releases/download/$version"
else
	base="$FORGELORE_BASE_URL"
fi

artifact="forgelore_${goos}_${goarch}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "==> downloading $artifact"
curl -fsSL -o "$tmp/$artifact" "$base/$artifact" || fail "could not download $base/$artifact"
curl -fsSL -o "$tmp/SHA256SUMS" "$base/SHA256SUMS" || fail "could not download the checksums"

echo "==> verifying"
expected="$(awk -v a="$artifact" '$2 == a || $2 == "*"a {print $1}' "$tmp/SHA256SUMS" | head -1)"
[ -n "$expected" ] || fail "SHA256SUMS does not mention $artifact"

if command -v sha256sum >/dev/null 2>&1; then
	actual="$(sha256sum "$tmp/$artifact" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
	actual="$(shasum -a 256 "$tmp/$artifact" | awk '{print $1}')"
else
	fail "neither sha256sum nor shasum is installed, so the download cannot be verified"
fi

if [ "$expected" != "$actual" ]; then
	fail "checksum mismatch for $artifact
  expected $expected
  got      $actual
Nothing was installed."
fi

mkdir -p "$install_dir"
chmod +x "$tmp/$artifact"
mv "$tmp/$artifact" "$install_dir/forgelore"
echo "==> installed $install_dir/forgelore"
"$install_dir/forgelore" version

# Installing somewhere the shell will not look is the commonest way this
# goes wrong, and it is silent unless somebody says so.
case ":$PATH:" in
*":$install_dir:"*) ;;
*)
	echo
	echo "$install_dir is not on your PATH. Add it:"
	echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
	echo "Then open a new shell, or run: hash -r"
	;;
esac
