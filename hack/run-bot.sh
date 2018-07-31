#!/usr/bin/env bash

set -e

cd $(dirname $0)/..

make cherry_pick_bot
./cherry_pick_bot
