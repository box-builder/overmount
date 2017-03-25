#!/bin/sh

set -e

if [ ! -x bin/box ]
then
  mkdir -p bin
  curl -sSL https://github.com/box-builder/box/releases/download/v0.5.1/box-0.5.1.linux.gz | gzip -dc >bin/box
  chmod +x bin/box
fi
