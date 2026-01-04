# GOTH-HW: Salesforce Browser

A GOTH stack (Go, Templ, HTMX, Tailwind) application for browsing Salesforce objects.

## Features

- Browse Accounts, Contacts, Opportunities, and Leads
- Column-based filtering with debounced search
- Paginated results
- Direct links to open records in Salesforce

## Prerequisites

- Go 1.23+
- Node.js (for Tailwind CSS)
- [templ](https://templ.guide/) CLI: `go install github.com/a-h/templ/cmd/templ@latest`
- Salesforce Connected App with Client Credentials flow enabled

## Salesforce Setup

1. Create a Connected App in Salesforce Setup
2. Enable OAuth settings with "Client Credentials Flow"
3. Add appropriate scopes (e.g., `api`, `refresh_token`)
4. Note the Consumer Key (Client ID) and Consumer Secret

## Setup

```bash
# Install Go dependencies
go mod tidy

# Install Tailwind
npm install

# Generate Templ files
templ generate

# Build Tailwind CSS
npm run css

# Copy .env.example to .env and fill in your credentials
cp .env.example .env
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SF_CLIENT_ID` | Salesforce Connected App Consumer Key |
| `SF_CLIENT_SECRET` | Salesforce Connected App Consumer Secret |
| `SF_LOGIN_URL` | `https://login.salesforce.com` or `https://test.salesforce.com` |
| `PORT` | Server port (default: 8080) |

## Running

```bash
# Load environment variables and run
source .env && go run ./cmd/server
```

## Development

Using [Air](https://github.com/air-verse/air) for hot reload:

```bash
go run github.com/air-verse/air@latest
```

Air handles everything: templ generation, Tailwind CSS build, and Go compilation on each change.

> **Note:** If your environment variables aren't in your shell profile (e.g., `.zshenv`), copy `.env.example` to `.env` and run `source .env` first.

## Project Structure

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Environment configuration
│   ├── handlers/        # HTTP handlers
│   └── salesforce/      # Salesforce API client
├── templates/           # Templ templates
├── static/              # Static assets (CSS)
└── tailwind.config.js   # Tailwind configuration
```
