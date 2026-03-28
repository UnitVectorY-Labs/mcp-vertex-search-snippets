
# Commands for mcp-vertex-search-snippets
default:
  @just --list
# Build mcp-vertex-search-snippets with Go
build:
  go build ./...

# Run tests for mcp-vertex-search-snippets with Go
test:
  go clean -testcache
  go test ./...