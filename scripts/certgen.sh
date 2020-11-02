#!/usr/bin/env bash

mkcert -cert-file certs/local.pem -key-file certs/local.key.pem localhost 127.0.0.1 ::1
#MKCERT_ROOT_DIR="$(printf %q "$(mkcert -CAROOT)")"
MKCERT_ROOT_DIR=$(mkcert -CAROOT)
sudo chown -R "$(id -un)": "$MKCERT_ROOT_DIR"
echo "copying CA Root from $MKCERT_ROOT_DIR"
cp -v "$MKCERT_ROOT_DIR"/rootCA.pem certs/
mkcert -install

