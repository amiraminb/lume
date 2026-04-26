#!/usr/bin/env bash
set -euo pipefail

REPO="amiraminb/lume"
EXT_DIR="${HOME}/.config/timewarrior/extensions"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "${arch}" in
  x86_64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported architecture: ${arch}" >&2; exit 1 ;;
esac

case "${os}" in
  darwin|linux) ;;
  *) echo "Unsupported OS: ${os}" >&2; exit 1 ;;
esac

version="${LUME_VERSION:-}"
if [[ -z "${version}" ]]; then
  version="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -o '"tag_name": "[^"]*"' \
    | head -n1 \
    | cut -d'"' -f4)"
fi

if [[ -z "${version}" ]]; then
  echo "Could not determine latest version" >&2
  exit 1
fi

version_stripped="${version#v}"
archive="lume_${version_stripped}_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${version}/${archive}"

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

echo "Downloading ${url}"
curl -fsSL "${url}" -o "${tmp}/${archive}"
tar -xzf "${tmp}/${archive}" -C "${tmp}"

mkdir -p "${EXT_DIR}"
install -m 0755 "${tmp}/lume" "${EXT_DIR}/lume"

echo "Installed lume ${version} to ${EXT_DIR}/lume"
