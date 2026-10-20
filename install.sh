#!/bin/sh
set -eu

repo="savchenko-dev/ccf"
dest="${CCF_INSTALL_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin|linux) ;;
  *) echo "ccf: unsupported OS '$os' (on Windows use install.ps1)" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "ccf: unsupported architecture '$arch'" >&2; exit 1 ;;
esac

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

url="https://github.com/$repo/releases/latest/download/ccf_${os}_${arch}.tar.gz"
echo "Downloading $url"
curl -fsSL "$url" -o "$tmp/ccf.tar.gz"
tar -xzf "$tmp/ccf.tar.gz" -C "$tmp"

mkdir -p "$dest"
install -m 755 "$tmp/ccf" "$dest/ccf"
echo "Installed ccf to $dest/ccf"

case ":$PATH:" in
  *":$dest:"*) ;;
  *) echo "Add $dest to your PATH to use ccf." ;;
esac
