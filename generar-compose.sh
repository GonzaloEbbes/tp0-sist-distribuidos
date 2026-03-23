#!/bin/bash
set -eu

DATASET_DIR=".data"
DATASET_ZIP=".data/dataset.zip"

if [ ! -f "$DATASET_DIR/agency-1.csv" ]; then
  mkdir -p "$DATASET_DIR"
  unzip -q "$DATASET_ZIP" -d "$DATASET_DIR"
fi

go run tools/composer.go "$@"
