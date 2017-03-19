#!/bin/sh

set -e

if [ "x$(which box)" = "x" ]
then
  echo "Installing erikh/box to build docker images; may require sudo password."
  curl -sSL box-builder.sh | bash
fi
