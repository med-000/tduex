#!/usr/bin/env sh
set -eu

BIN_NAME="tduex"
DEFAULT_USER_BIN="${HOME}/.local/bin"

choose_bin_dir() {
	if [ -n "${BIN_DIR:-}" ]; then
		printf '%s\n' "$BIN_DIR"
		return
	fi

	if command -v go >/dev/null 2>&1; then
		GOBIN_VALUE="$(go env GOBIN 2>/dev/null || true)"
		if [ -n "$GOBIN_VALUE" ]; then
			printf '%s\n' "$GOBIN_VALUE"
			return
		fi

		GOPATH_VALUE="$(go env GOPATH 2>/dev/null || true)"
		if [ -n "$GOPATH_VALUE" ] && [ -w "$GOPATH_VALUE/bin" ] 2>/dev/null; then
			printf '%s\n' "$GOPATH_VALUE/bin"
			return
		fi
	fi

	if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
		printf '%s\n' "/usr/local/bin"
		return
	fi

	printf '%s\n' "$DEFAULT_USER_BIN"
}

BIN_DIR="$(choose_bin_dir)"

mkdir -p "$BIN_DIR"
go build -o "$BIN_DIR/$BIN_NAME" ./cmd/tduex

printf 'installed %s\n' "$BIN_DIR/$BIN_NAME"

case ":$PATH:" in
	*:"$BIN_DIR":*)
		;;
	*)
		printf 'add this to your shell profile:\n  export PATH="%s:$PATH"\n' "$BIN_DIR"
		;;
esac
