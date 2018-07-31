#!/usr/bin/env bash

cd $(dirname $0)/..

make cherry_pick_bot
./cherry_pick_bot
