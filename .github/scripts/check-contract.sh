#!/usr/bin/env bash
set -euo pipefail

latest_tag="$(git tag --list 'api-contract-v1.*' --sort=-v:refname | head -n 1)"
if [[ -z "${latest_tag}" ]]; then
  echo "No released api-contract-v1.* tag exists; this is the initial contract baseline."
  exit 0
fi

baseline="$(mktemp)"
trap 'rm -f "${baseline}"' EXIT
git show "${latest_tag}:openapi.yaml" > "${baseline}"

echo "Checking openapi.yaml against ${latest_tag}."
go tool oasdiff breaking "${baseline}" openapi.yaml --fail-on ERR

