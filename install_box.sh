#!/bin/sh

set -e

if [ "x$(which box)" = "x" ]
then
  echo "Installing erikh/box to build docker images; may require sudo password."
  curl -sSL https://raw.githubusercontent.com/erikh/box/master/install.sh | sudo bash
fi
