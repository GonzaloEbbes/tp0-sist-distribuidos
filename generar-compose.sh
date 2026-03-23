#!/bin/bash
set -eu

DATASET_DIR=".data/dataset"
DATASET_ZIP=".data/dataset.zip"

if [ ! -d "$DATASET_DIR" ]; then
  mkdir -p "$DATASET_DIR"
  unzip -q "$DATASET_ZIP" -d "$DATASET_DIR"
fi

go run tools/composer.go "$@"
