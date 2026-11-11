#!/usr/bin/env bash
# setup-golangci-lint helper, called by action.yml (why: see there).
#   setup.sh download   fetch the tarball into $RUNNER_TEMP/golangci-lint-dl, verify its attestation
#   setup.sh install    check sha256 (cache hit or miss), extract, put the binary on $GITHUB_PATH
# Env: VERSION (2.15.0), SHA256 (install), GH_TOKEN (download), RUNNER_TEMP, GITHUB_PATH.
# SETUP_GL_BACKOFF: seconds per retry step (default 15; tests set 0).
set -euo pipefail
{
mode="${1:-}"
: "${VERSION:?VERSION is required}" "${RUNNER_TEMP:?RUNNER_TEMP is required}"
dir="$RUNNER_TEMP/golangci-lint-dl"
name="golangci-lint-${VERSION}-linux-amd64"
f="$name.tar.gz"
backoff="${SETUP_GL_BACKOFF:-15}"

case "$mode" in
download)
  mkdir -p "$dir"
  ok=
  for attempt in 1 2 3 4; do
    rm -f "$dir/$f"
    # The API asset route served the file on 2026-11-08 while the CDN URL returned 504.
    if gh release download "v${VERSION}" -R golangci/golangci-lint -p "$f" -D "$dir"; then ok=1; break; fi
    if curl -fsSL --retry 3 --retry-all-errors --retry-delay 5 -o "$dir/$f" \
        "https://github.com/golangci/golangci-lint/releases/download/v${VERSION}/$f"; then ok=1; break; fi
    echo "attempt $attempt failed on both routes"
    sleep $((attempt * backoff))
  done
  if [ -z "$ok" ]; then echo "could not download $f"; exit 1; fi
  # Proves golangci's release workflow built these bytes, not only that they match a hash.
  gh attestation verify "$dir/$f" --repo golangci/golangci-lint
  ;;
install)
  : "${SHA256:?SHA256 is required}" "${GITHUB_PATH:?GITHUB_PATH is required}"
  # On a cache hit too: the pin decides, never what the cache holds.
  echo "${SHA256}  $dir/$f" | sha256sum -c -
  bin="$RUNNER_TEMP/golangci-lint-bin"
  mkdir -p "$bin"
  tar -xzf "$dir/$f" -C "$bin" --strip-components=1 "$name/golangci-lint"
  echo "$bin" >> "$GITHUB_PATH"
  ;;
*)
  echo "usage: setup.sh download|install"; exit 2
  ;;
esac
exit 0
}
