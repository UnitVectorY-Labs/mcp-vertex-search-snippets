
# mcp-vertex-search-snippets

A lightweight MCP server that integrates with Vertex AI Search to retrieve configurable snippets and extractive segments for document discovery.

## Purpose

mcp-vertex-search-snippets allows you to expose Vertex AI Search as an MCP tool.
It provides an MCP-compliant search tool that accepts a query string returning high-quality search snippets or extractive results from the files indexed by Vertex AI Search.

## Releases

Versions of `mcp-vertex-search-snippets` are published on GitHub Releases. Since this MCP server is written in Go, each release provides pre-compiled executables for macOS, Linux, and Windows—ready to download and run.

Alternatively, if you have Go installed, you can install mcp-vertex-search-snippets directly from source using:

```bash
go install github.com/UnitVectorY-Labs/mcp-vertex-search-snippets@latest
```

## Configuration

The server is configured using command line parameters, environment variables, and a YAML configuration file (vertex.yaml).

### Command Line Parameters

- `--vertexConfig`: Path to the configuration YAML file. If set, this takes precedence over the VERTEX_CONFIG environment variable. If neither is provided, the application exits with an error.
- `--vertexDebug`: If provided, enables detailed debug logging to stderr, including HTTP request/response dumps. If set, this takes precedence over the VERTEX_DEBUG environment variable.
- `--http`: Run the server in streamable HTTP mode on the given port (e.g., --http 8080). Defaults to stdio.

### Environment Variables

- `VERTEX_CONFIG`: Path to the configuration YAML file. Used if --vertexConfig is not set.
- `VERTEX_DEBUG`: If set to true (case-insensitive), enables detailed debug logging. Used if --vertexDebug is not set.

### vertex.yaml

The vertex.yaml file specifies the Discovery Engine configuration:

```yaml
project_id: "000000000000"
location: "us"
app_id: "example_0000000000000"
```

Attributes:
- `project_id`: Your Google Cloud project ID.
- `location`: One of global, us, or eu.
- `app_id`: The Discovery Engine app/engine identifier.

## MCP Tools

This MCP server exposes a single tool:

### search

Search for relevant documents based on the provided query.

Inputs:
- `query` (string, required): The search text.
- `maxExtractiveSegmentCount` (number, optional): Maximum number of extractive segments to return (default: 1).

Annotations:
- `title`: "Search"
- `readOnlyHint`: true
- `destructiveHint`: false
- `idempotentHint`: true
- `openWorldHint`: true

## Run in Streamable HTTP Mode

By default, the server runs in stdio mode. To run in streamable HTTP mode:

```bash
./mcp-vertex-search-snippets --http 8080
```

Your MCP client can then connect to:

http://localhost:8080/mcp

If an Authorization header is passed to the MCP server, it is used for authentication with Vertex AI Search. Otherwise, the server obtains credentials via Google Cloud’s default authentication chain (service accounts, gcloud CLI, etc.).

## Limitations

- Each instance of mcp-vertex-search-snippets can only connect to a single Discovery Engine instance defined in vertex.yaml.
- Only the search tool is currently exposed; resources are not exposed as MCP Resources.
- Requires valid Google Cloud credentials available via Application Default Credentials.
