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
- Web browser with JavaScript enabled

## Installation

1. Clone the repository
```bash
git clone https://github.com/patelrohanv/alt-text-generator.git
cd alt-text-generator
```

2. Install dependencies
```bash
go mod tidy
```

3. Create a .env file in the root directory and add your API key(s):
```env
OPEN_AI_API_KEY=your_openai_key_here
ANTHROPIC_API_KEY=your_anthropic_key_here
```

## Building and Running

First, make the build script executable:
```bash
chmod +x build.sh
```

Then you can use it in several ways:

```bash
# Simple build
./build.sh

# Build with specific mode (creates a run script)
./build.sh -m openai
./build.sh -m anthropic

# Build with debug symbols
./build.sh -d

# Show build script help
./build.sh -h
```

After building, run the application:

```bash
# If built with -m flag:
./bin/run.sh

# Otherwise:
./bin/alt-text-generator -openai
# or
./bin/alt-text-generator -anthropic
```

Then open your web browser and navigate to:
```
http://localhost:8080
```

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
│   └── types/
│       └── types.go
├── web/
│   └── template.html
├── build.sh
└── .env
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

