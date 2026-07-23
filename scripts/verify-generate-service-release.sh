#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

[[ -z "$(git status --porcelain)" ]] || {
  echo "release verification refused: checkout is dirty" >&2
  exit 1
}

source_commit="$(git rev-parse HEAD)"
source_tag="$(git tag --points-at HEAD | awk '/^v[0-9]+\.[0-9]+\.[0-9]+$/ { print }')"
[[ "$(printf '%s\n' "$source_tag" | sed '/^$/d' | wc -l | tr -d ' ')" == "1" ]] || {
  echo "release verification refused: HEAD must have exactly one semver release tag" >&2
  exit 1
}
git merge-base --is-ancestor HEAD origin/main || {
  echo "release verification refused: tagged commit is not reachable from fetched origin/main" >&2
  exit 1
}

declared_version="$(sed -n 's/^const Version = "\([0-9][0-9.]*\)"$/\1/p' internal/version/version.go)"
[[ "v$declared_version" == "$source_tag" ]] || {
  echo "release verification refused: source version $declared_version does not match $source_tag" >&2
  exit 1
}

go test ./...
go test ./gen -run '^TestV7KnownVector$' -count=1
docker build \
  --build-arg "SOURCE_COMMIT=$source_commit" \
  --build-arg "SOURCE_TAG=$source_tag" \
  --file cmd/generate-service/Dockerfile \
  --tag "generate-service:$source_tag" \
  .
