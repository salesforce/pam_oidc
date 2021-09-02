#!/usr/bin/env bash
set -euo pipefail

if [[ "${GITHUB_REF-}" =~ ^refs/tags/ ]]; then
  echo "${GITHUB_REF#refs/tags/}"
  exit
fi

if [[ -n "${GITHUB_SHA-}" ]]; then
  echo "${GITHUB_SHA}"
  exit
fi

echo "dev"
