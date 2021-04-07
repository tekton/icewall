#!/bin/bash

branch=`git branch --show-current`
rev=`git rev-parse --short HEAD`

echo "{\"build rev\": \"${rev}\", \"branch\": \"${branch}\"}" > icewall_build_meta.json