#!/usr/bin/env bash
set -uex -o pipefail

log() {
    echo -e "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] $1" >&2
}

GO_VERSION=1.12.4
GO_INSTALL_DIR=${HOME}/go_installs/${GO_VERSION}
if [[ ! -d ${GO_INSTALL_DIR}/go ]]; then
    mkdir -p ${GO_INSTALL_DIR}
    curl -Sfl https://dl.google.com/go/go1.12.4.linux-amd64.tar.gz |tar -xz -C ${GO_INSTALL_DIR}
fi


export GOPATH="${HOME}/go"
export PATH="${GO_INSTALL_DIR}/go/bin:${GOPATH}/bin:${PATH}"

log "Building..."

make release
make | tee build.log
