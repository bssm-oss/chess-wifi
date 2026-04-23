#!/usr/bin/env sh
set -eu

repo="github.com/bssm-oss/chess-wifi/cmd/chess-wifi@latest"
binary="chess-wifi"
tmpdir="$(mktemp -d)"

cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

if ! command -v go >/dev/null 2>&1; then
  echo "go is required to install chess-wifi." >&2
  exit 1
fi

echo "Building $repo"
GOBIN="$tmpdir" go install "$repo"

install_dir="${CHESS_WIFI_INSTALL_DIR:-/usr/local/bin}"
target="$install_dir/$binary"

mkdir_with_sudo() {
  if [ -d "$install_dir" ]; then
    return 0
  fi
  if mkdir -p "$install_dir" 2>/dev/null; then
    return 0
  fi
  if command -v sudo >/dev/null 2>&1; then
    sudo mkdir -p "$install_dir"
    return 0
  fi
  return 1
}

install_binary() {
  if install -m 0755 "$tmpdir/$binary" "$target" 2>/dev/null; then
    return 0
  fi
  if command -v sudo >/dev/null 2>&1; then
    sudo install -m 0755 "$tmpdir/$binary" "$target"
    return 0
  fi
  return 1
}

if ! mkdir_with_sudo || ! install_binary; then
  fallback="${HOME}/.local/bin"
  mkdir -p "$fallback"
  install -m 0755 "$tmpdir/$binary" "$fallback/$binary"
  target="$fallback/$binary"
fi

echo "Installed $binary to $target"

case ":$PATH:" in
  *":$(dirname "$target"):"*)
    run_cmd="$binary match"
    ;;
  *)
    echo "$(dirname "$target") is not in PATH. Add it to your shell profile or set CHESS_WIFI_INSTALL_DIR to a PATH directory." >&2
    run_cmd="$target match"
    ;;
esac

"$target" --help >/dev/null
echo "Run: $run_cmd"
