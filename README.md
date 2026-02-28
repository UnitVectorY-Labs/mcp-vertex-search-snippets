[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/mcp-vertex-search-snippets.svg)](https://github.com/UnitVectorY-Labs/mcp-vertex-search-snippets/releases/latest) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active) [![Go Report Card](https://goreportcard.com/badge/github.com/UnitVectorY-Labs/mcp-vertex-search-snippets)](https://goreportcard.com/report/github.com/UnitVectorY-Labs/mcp-vertex-search-snippets)

# mcp-vertex-search-snippets

A lightweight [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server that integrates with [Vertex AI Search](https://cloud.google.com/enterprise-search) (Discovery Engine) to retrieve snippets and extractive segments for document discovery.

## Overview

mcp-vertex-search-snippets bridges the gap between AI assistants (such as GitHub Copilot, Claude Desktop, or any MCP-compatible client) and your organization's documents indexed in Google Cloud's Vertex AI Search. It exposes a single MCP `search` tool that an AI assistant can invoke to query your Discovery Engine data store and receive relevant document excerpts.

### Use Case

If your organization has internal documentation, knowledge bases, or website content indexed in Vertex AI Search, this server allows AI assistants to search that content in real time. For example:

- An AI coding assistant can look up internal API documentation or runbooks while helping a developer.
- A support chatbot can search a company knowledge base for relevant articles to answer customer questions.
- A research assistant can query indexed papers or reports for relevant passages.

The server returns results in a priority order: **extractive segments** (longer, contextual passages) are preferred, followed by **snippets** (shorter highlighted excerpts), and finally **document title and link** as a fallback. This ensures the AI assistant receives the most useful content available.

### How It Works

```
MCP Client (e.g., VS Code, Claude Desktop)
    |
    |  MCP Protocol (stdio or HTTP)
    v
mcp-vertex-search-snippets
    |
    |  REST API (authenticated)
    v
Vertex AI Search (Discovery Engine)
    |
    v
Your Indexed Documents
```

1. An MCP client sends a `search` tool call with a query string.
2. The server constructs a request to the [Discovery Engine `servingConfigs.search`](https://cloud.google.com/generative-ai-app-builder/docs/reference/rest/v1/projects.locations.collections.engines.servingConfigs/search) REST API, requesting both snippets and extractive segments.
3. The API response is parsed and the most relevant content is returned as plain text to the MCP client.

## Prerequisites

- **Google Cloud Project** with [Vertex AI Search](https://cloud.google.com/enterprise-search) enabled.
- **Discovery Engine App** configured with a data store containing your indexed content (unstructured documents or website data).
- **Authentication** via one of:
  - [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) (ADC) -- the default when running locally or on Google Cloud infrastructure.
  - An `Authorization` header passed through the MCP HTTP transport.
- The authenticated principal needs the **Discovery Engine Viewer** (`roles/discoveryengine.viewer`) role or equivalent permissions on the Discovery Engine resource.

## Installation

Versions of `mcp-vertex-search-snippets` are published on [GitHub Releases](https://github.com/UnitVectorY-Labs/mcp-vertex-search-snippets/releases/latest). Each release provides pre-compiled executables for macOS, Linux, and Windows ready to download and run.

Alternatively, if you have Go installed, you can install directly from source:

```bash
go install github.com/UnitVectorY-Labs/mcp-vertex-search-snippets@latest
```

## Configuration

The server is configured using command line flags, environment variables, and a YAML configuration file.

### Command Line Flags

| Flag | Description |
|---|---|
| `--vertexConfig` | Path to the configuration YAML file. Overrides the `VERTEX_CONFIG` environment variable. |
| `--vertexDebug` | Enable detailed debug logging to stderr, including HTTP request/response dumps. Overrides the `VERTEX_DEBUG` environment variable. |
| `--http` | Run in streamable HTTP mode on the given port (e.g., `--http 8080`). Defaults to stdio transport. |

### Environment Variables

| Variable | Description |
|---|---|
| `VERTEX_CONFIG` | Path to the configuration YAML file. Used when `--vertexConfig` is not set. |
| `VERTEX_DEBUG` | Set to `true` to enable debug logging. Used when `--vertexDebug` is not set. |

### Configuration File (vertex.yaml)

The YAML configuration file specifies which Discovery Engine app to query:

```yaml
project_id: "000000000000"
location: "us"
app_id: "example_0000000000000"
```

| Field | Description |
|---|---|
| `project_id` | Your Google Cloud project number or ID. |
| `location` | The location of your Discovery Engine app. Must be one of `global`, `us`, or `eu`. This determines both the API endpoint and the resource path. |
| `app_id` | The Discovery Engine app (engine) identifier. |

The `location` value determines the API endpoint used:

| Location | API Endpoint |
|---|---|
| `global` | `https://discoveryengine.googleapis.com` |
| `us` | `https://us-discoveryengine.googleapis.com` |
| `eu` | `https://eu-discoveryengine.googleapis.com` |

## MCP Tool

This server exposes a single tool:

### search

Search for relevant documents based on the provided query.

**Inputs:**

| Parameter | Type | Required | Description |
|---|---|---|---|
| `query` | string | Yes | The search text. |
| `maxExtractiveSegmentCount` | number | No | Maximum number of extractive segments to return per document. Defaults to `1`. Must be at least `1`. |

**Annotations:**

| Annotation | Value | Description |
|---|---|---|
| `title` | Search | Display name for the tool. |
| `readOnlyHint` | true | The tool does not modify any data. |
| `destructiveHint` | false | The tool is not destructive. |
| `idempotentHint` | true | Repeated calls with the same input produce the same result. |
| `openWorldHint` | true | The tool interacts with an external service. |

**Response Format:**

The tool returns plain text with results separated by `---`. For each search result, content is selected in priority order:

1. **Extractive segments** -- longer, contextual passages extracted directly from the document.
2. **Snippets** -- shorter highlighted excerpts matching the query.
3. **Title and link** -- the document title and URL as a fallback when no content is available.

## Transport Modes

### Stdio (Default)

By default, the server communicates over stdin/stdout using the MCP stdio transport. This is the standard mode for local MCP integrations.

```bash
./mcp-vertex-search-snippets --vertexConfig vertex.yaml
```

### Streamable HTTP

To run the server as an HTTP endpoint:

```bash
./mcp-vertex-search-snippets --vertexConfig vertex.yaml --http 8080
```

The MCP endpoint is available at `http://localhost:8080/mcp`.

When running in HTTP mode, the server checks for an `Authorization` header on incoming requests. If present, the header value is forwarded directly to the Vertex AI Search API. If no `Authorization` header is provided, the server falls back to [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials).

## MCP Client Configuration Examples

### VS Code

Add the following to your VS Code settings (`.vscode/settings.json` or user settings):

```json
{
  "mcp": {
    "servers": {
      "vertex-search": {
        "command": "mcp-vertex-search-snippets",
        "args": [],
        "env": {
          "VERTEX_CONFIG": "/path/to/vertex.yaml"
        }
      }
    }
  }
}
```

### Claude Desktop

Add the following to your Claude Desktop configuration file:

```json
{
  "mcpServers": {
    "vertex-search": {
      "command": "mcp-vertex-search-snippets",
      "args": ["--vertexConfig", "/path/to/vertex.yaml"]
    }
  }
}
```

## Limitations

- Each instance connects to a single Discovery Engine app defined in the configuration file. To query multiple apps, run multiple instances.
- Only the `search` tool is exposed; indexed documents are not exposed as MCP Resources.
- Requires valid Google Cloud credentials available via Application Default Credentials or an Authorization header (HTTP mode only).
