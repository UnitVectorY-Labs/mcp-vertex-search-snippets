# Example Config

This example shows a minimal setup pointing to a Discovery Engine **engine** serving config.

## VS Code Test

```json
{
  "mcp": {
    "inputs": [],
    "servers": {
      "vertex": {
        "command": "mcp-vertex-search-snippets",
        "args": [],
        "env": {
          "VERTEX_CONFIG": "mcp-vertex-search-snippets/example/vertex.yaml"
        }
      }
    }
  }
}
```
