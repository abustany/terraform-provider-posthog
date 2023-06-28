#!/bin/sh

set -e

go generate ./...
git diff --compact-summary --exit-code || \
  (echo; echo "Unexpected difference in directories after code generation. Run 'go generate ./...' command and commit."; exit 1)
