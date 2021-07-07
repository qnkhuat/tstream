#!/bin/bash
# exit when any command fails
set -e

echo "Version: "
read version
git tag -a $version -m "Release version $version"
git push origin $version
goreleaser release --rm-dist
