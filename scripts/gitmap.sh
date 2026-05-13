#!/usr/bin/env bash
# gitmap shell integration — add to ~/.bashrc or ~/.zshrc:
#   source <(gitmap --shell-init)

# shellcheck disable=SC2120
gmap() {
    local target
    target=$(gitmap 2>/dev/null)
    if [ -n "$target" ] && [ -d "$target" ]; then
        cd "$target" || return
        echo "→ $target"
    fi
}

# Allow sourcing this file directly:
#   source ~/projects/gitmap/scripts/gitmap.sh

true
