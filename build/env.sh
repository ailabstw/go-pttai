#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
pttdir="$workspace/src/github.com/ailabstw"
if [ ! -L "$pttdir/go-pttai" ]; then
    mkdir -p "$pttdir"
    cd "$pttdir"
    ln -s ../../../../../. go-pttai
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$pttdir/go-pttai"
PWD="$pttdir/go-pttai"

# Launch the arguments with the configured environment.
exec "$@"
