# Anna's Archive MCP Server (and CLI Tool)

[An MCP server](https://modelcontextprotocol.io/introduction) and CLI tool for searching and downloading documents from [Anna's Archive](https://annas-archive.li)

> [!NOTE]
> Notwithstanding prevailing public sentiment regarding Anna's Archive, the platform serves as a comprehensive repository for automated retrieval of documents released under permissive licensing frameworks (including Creative Commons publications and public domain materials). This software does not endorse unauthorized acquisition of copyrighted content and should be regarded solely as a utility. Users are urged to respect the intellectual property rights of authors and acknowledge the considerable effort invested in document creation.

> [!WARNING]
> Please refer to [this section](#annas-archive-mirrors) if any of the links lead to a non-functional Anna's Archive server.

## Available Operations

| Operation                                      | MCP Tool           | CLI Command         |
| ---------------------------------------------- | ------------------ | ------------------- |
| Search for books by title, author, or topic   | `book_search`      | `book-search`       |
| Download a book by its MD5 hash                | `book_download`    | `book-download`     |
| Search for articles by DOI or keywords        | `article_search`   | `article-search`    |
| Download an article by its DOI                 | `article_download` | `article-download`  |

### Usage Examples

**Searching for Books**
- **MCP**: Use the `book_search` tool with a query like `"machine learning python"`
- **CLI**: `annas-mcp book-search "machine learning python"`

**Downloading Books**
- **MCP**: Use `book_download` with the hash from search results
- **CLI**: `annas-mcp book-download abc123def456 "my-book.pdf"`

**Searching for Articles**
- **MCP**: Use `article_search` with either a DOI or keywords
  - By DOI: `article_search` with query `"10.1038/nature12345"`
  - By keywords: `article_search` with query `"neural networks"`
- **CLI**:
  - By DOI: `annas-mcp article-search "10.1038/nature12345"`
  - By keywords: `annas-mcp article-search "neural networks"`

**Downloading Articles**
- **MCP**: Use `article_download` with the DOI
- **CLI**: `annas-mcp article-download "10.1038/nature12345"`

## Requirements

If you plan to use only the CLI tool, you need:

- [A donation to Anna's Archive](https://annas-archive.li/donate), which grants JSON API access
- [An API key](https://annas-archive.li/faq#api)

If using the project as an MCP server, you also need an MCP client, such as [Claude Desktop](https://claude.ai/download).

The environment should contain two variables:

- `ANNAS_SECRET_KEY`: The Anna's Archive API key.
- `ANNAS_DOWNLOAD_PATH`: The path where the documents should be downloaded.

Optionally, you can set:

- `ANNAS_BASE_URL`: The base URL of the Anna's Archive mirror to use (defaults to `annas-archive.li`).

These variables can also be stored in an `.env` file in the folder containing the binary.

## Setup

Download the appropriate binary from [the GitHub Releases section](https://github.com/iosifache/annas-mcp/releases).

If you plan to use the tool for its MCP server functionality, you need to integrate it into your MCP client. If you are using Claude Desktop, please consider the following example configuration:

```json
"anna-mcp": {
    "command": "/Users/iosifache/Downloads/annas-mcp",
    "args": ["mcp"],
    "env": {
        "ANNAS_SECRET_KEY": "feedfacecafebeef",
        "ANNAS_DOWNLOAD_PATH": "/Users/iosifache/Downloads"
    }
}
```

## Demo

### As an MCP Server

<img src="screenshots/claude.png" width="600px"/>

### As a CLI Tool

<img src="screenshots/cli.png" width="400px"/>

## Healthcheck

To verify that the installation is working correctly, run the healthcheck script:

```bash
./scripts/healthcheck.sh
```

This script tests:
- Book search functionality (searching for "crypto")
- Article search functionality (searching for a paper by DOI)

Both tests should pass if the service is functioning properly.

### Automated Healthchecks

The repository includes a GitHub Action (`.github/workflows/healthcheck.yml`) that runs daily at 00:00 UTC to verify the service is working. The workflow:
- Runs the healthcheck script automatically
- Reports failures if any test fails
- Can be triggered manually via the Actions tab

#### Manual Trigger

To manually run the healthcheck:
1. Go to the **Actions** tab in GitHub
2. Select **"Daily Healthcheck"** from the workflows list
3. Click **"Run workflow"**
4. Optionally provide a reason for the manual run
5. Click **"Run workflow"** to start

## Anna's Archive Mirrors

Anna's Archive has multiple mirrors, which may be innactive at times due to various reasons. Below is a list of known mirrors and their status as of January 2025:

| Mirror                                           | Type     | Status    |
| ------------------------------------------------ | -------- | --------- |
| [`annas-archive.li`](https://annas-archive.li)   | Official | Active    |
| [`annas-archive.pm`](https://annas-archive.pm)   | Official | Active    |
| [`annas-archive.in`](https://annas-archive.in)   | Official | Active    |
| [`annas-archive.org`](https://annas-archive.org) | Official | Innactive |

Alternatively, use [The Shadow Library Uptime Monitor](https://open-slum.org) to find statuses or alternative mirrors.

This project defaults to `annas-archive.li`. If that mirror is not working for you, please set the `ANNAS_BASE_URL` environment variable to one of the other mirrors.
