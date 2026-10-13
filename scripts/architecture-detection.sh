#!/usr/bin/env bash
set -o errtrace -o pipefail -o noclobber -o errexit -o nounset

APP="architecture-detection"

exit_error() {
  echo "${APP} error: ${1}" >&2
  exit "${2:-1}"
}

# arm64 on macos, aarch64 on linux
case $(uname -m) in
  x86_64)        echo "amd64" ;;
  arm64|aarch64) echo "arm64" ;;
  *)      exit_error "unsupported architecture $(uname -m)"
esac
