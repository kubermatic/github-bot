#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

make github-bot
./github-bot -logtostderr -v=6
