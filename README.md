# Alt Text Generator

A web application that generates alt text descriptions for images using either OpenAI's GPT or Anthropic's Claude API.

## Features

- Support for both OpenAI and Claude APIs
- Simple web interface for image uploads
- Client-side file size validation
- Secure API key management
- Support for JPG, PNG, and GIF formats
- Maximum file size: 5MB

## Prerequisites

- Go 1.21 or higher
- OpenAI API key or Anthropic API key
- templ compiler
- Web browser with JavaScript enabled

## Installation

1. Clone the repository
```bash
git clone [your-repo-url]
cd alt-text-generator
```

2. Install templ compiler
```bash
go install github.com/a-h/templ/cmd/templ@latest
```

3. Install dependencies
```bash
go mod tidy
```

4. Generate templ files
```bash
templ generate
```

5. Create a .env file in the root directory and add your API key(s):
```env
OPEN_AI_API_KEY=your_openai_key_here
ANTHROPIC_API_KEY=your_anthropic_key_here
```

## Usage

### Using the Build Script

1. Make the build script executable:
```bash
chmod +x build.sh
```

2. Build and run with OpenAI:
```bash
./build.sh -m openai
./bin/run.sh
```

Or with Claude:
```bash
./build.sh -m anthropic
./bin/run.sh
```

### Manual Build and Run

1. Generate templ files:
```bash
templ generate
```

2. Run with OpenAI:
```bash
go run cmd/server/main.go -openai
```

Or with Claude:
```bash
go run cmd/server/main.go -anthropic
```

The server will start at http://localhost:8080

## Directory Structure

```
alt-text-generator/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── claude.go
│   │   └── openai.go
│   ├── config/
│   │   └── env.go
│   ├── handlers/
│   │   ├── home.go
│   │   ├── upload.go
│   │   └── apikey.go
│   ├── components/
│   │   ├── layout.templ
│   │   ├── home.templ
│   │   ├── upload.templ
│   │   └── alttext.templ
│   └── types/
│       └── types.go
├── build.sh                # Optional build script
├── .env                    # Your API keys (create this file)
└── go.mod
```

## Development

After making changes to any .templ files, regenerate the templates:
```bash
templ generate
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
