# GOTH-HW: Salesforce Browser

A GOTH stack (Go, Templ, HTMX, Tailwind) application for browsing Salesforce objects.

## Features

- Browse Accounts, Contacts, Opportunities, and Leads
- Column-based filtering with debounced search
- Paginated results
- Direct links to open records in Salesforce

## Prerequisites

- Git
- Go 1.23+
- Node.js 18+ (for Tailwind CSS)
- [templ](https://templ.guide/) CLI
- Salesforce Connected App with Client Credentials flow enabled

### Installing Prerequisites on a New Laptop

#### 1. Check What's Already Installed

```bash
# Check Git
git --version

# Check Go
go version

# Check Node.js
node --version
```

#### 2. Install Missing Tools

**Git:**
- macOS: `brew install git` or download from [git-scm.com](https://git-scm.com/)
- Linux: `sudo apt-get install git` or `sudo yum install git`
- Windows: Download from [git-scm.com](https://git-scm.com/)

**Go:**
- Download and install from [go.dev/dl](https://go.dev/dl/)
- Ensure Go's bin directory is in your PATH:
  ```bash
  # Add to ~/.zshrc or ~/.bashrc if not present
  export PATH=$PATH:$(go env GOPATH)/bin
  ```
- Reload your shell: `source ~/.zshrc` (or `~/.bashrc`)

**Node.js:**
- macOS: `brew install node` or use [nvm](https://github.com/nvm-sh/nvm)
- Linux: Use [nvm](https://github.com/nvm-sh/nvm) or your package manager
- Windows: Download from [nodejs.org](https://nodejs.org/) or use [nvm-windows](https://github.com/coreybutler/nvm-windows)

**templ CLI:**
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

#### 3. Verify Installation

```bash
# All of these should output version numbers
git --version
go version
node --version
templ --version
```

## Salesforce Setup

1. Create a Connected App in Salesforce Setup
2. Enable OAuth settings with "Client Credentials Flow"
3. Add appropriate scopes (e.g., `api`, `refresh_token`)
4. Note the Consumer Key (Client ID) and Consumer Secret

## Setup

Once all prerequisites are installed, set up the project:

```bash
# Clone the repository (if not already done)
git clone <repository-url>
cd goth-hw

# Install Go dependencies
go mod tidy

# Install Node.js dependencies (Tailwind CSS)
npm install

# Install templ CLI (if not done in prerequisites)
go install github.com/a-h/templ/cmd/templ@latest

# Generate Templ files (compiles .templ templates to Go)
templ generate

# Build Tailwind CSS (processes styles)
npm run css

# Copy environment template and configure
cp .env.example .env
# Edit .env with your Salesforce credentials
```

After running these commands, edit `.env` with your actual Salesforce credentials from the Connected App setup.

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

## Troubleshooting

### `templ: command not found`
Ensure `$(go env GOPATH)/bin` is in your PATH:
```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### `npm: command not found`
Node.js is not installed. Follow the Node.js installation steps in Prerequisites.

### Blank page or missing styles
Make sure you've run both:
```bash
templ generate  # Generates Go code from .templ files
npm run css     # Builds Tailwind CSS
```

### Environment variables not loading
If using Air for development, either:
- Add variables to your shell profile (`~/.zshrc` or `~/.bashrc`)
- Or run `source .env` before starting Air

### Port already in use
Change the `PORT` variable in your `.env` file to an unused port (e.g., 8081).
