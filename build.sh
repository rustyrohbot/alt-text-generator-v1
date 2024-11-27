#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Ensure script fails on any error
set -e

# Default values
OUTPUT_DIR="bin"
MODE=""
DEBUG=false

# Print usage information
usage() {
    echo -e "${BLUE}Usage: $0 [-m mode] [-d] [-h]${NC}"
    echo -e "  -m mode   : Build mode (openai or anthropic)"
    echo -e "  -d        : Enable debug build"
    echo -e "  -h        : Show this help message"
    exit 1
}

# Parse command line arguments
while getopts "m:dh" opt; do
    case $opt in
        m)
            MODE=$OPTARG
            if [[ "$MODE" != "openai" && "$MODE" != "anthropic" ]]; then
                echo -e "${RED}Error: Mode must be either 'openai' or 'anthropic'${NC}"
                exit 1
            fi
            ;;
        d)
            DEBUG=true
            ;;
        h)
            usage
            ;;
        \?)
            echo -e "${RED}Invalid option: -$OPTARG${NC}"
            usage
            ;;
    esac
done

# Verify required tools are installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

if ! command -v templ &> /dev/null; then
    echo -e "${RED}Error: templ is not installed${NC}"
    echo -e "${BLUE}Install with: go install github.com/a-h/templ/cmd/templ@latest${NC}"
    exit 1
fi

echo -e "${BLUE}Starting build process...${NC}"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build binary name
BINARY_NAME="alt-text-generator"
if [ "$DEBUG" = true ]; then
    BINARY_NAME="${BINARY_NAME}-debug"
fi

# Build flags
BUILD_FLAGS=""
if [ "$DEBUG" = false ]; then
    BUILD_FLAGS="-ldflags=-w -ldflags=-s"
fi

# Generate templ files
echo -e "${BLUE}Generating templ files...${NC}"
templ generate

echo -e "${BLUE}Running go mod tidy...${NC}"
go mod tidy

echo -e "${BLUE}Building binary...${NC}"
if go build $BUILD_FLAGS -o "$OUTPUT_DIR/$BINARY_NAME" cmd/server/main.go; then
    echo -e "${GREEN}Build successful!${NC}"
    echo -e "Binary location: $OUTPUT_DIR/$BINARY_NAME"
    
    # Create run script
    if [ ! -z "$MODE" ]; then
        RUN_SCRIPT="$OUTPUT_DIR/run.sh"
        echo "#!/bin/bash" > "$RUN_SCRIPT"
        echo "./$BINARY_NAME -$MODE" >> "$RUN_SCRIPT"
        chmod +x "$RUN_SCRIPT"
        echo -e "${GREEN}Created run script: $RUN_SCRIPT${NC}"
    fi
else
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

# Print usage instructions
echo -e "\n${BLUE}Usage instructions:${NC}"
if [ ! -z "$MODE" ]; then
    echo -e "Run the application with: ${GREEN}./bin/run.sh${NC}"
else
    echo -e "Run the application with: ${GREEN}./bin/$BINARY_NAME -openai${NC}"
    echo -e "                     or: ${GREEN}./bin/$BINARY_NAME -anthropic${NC}"
fi