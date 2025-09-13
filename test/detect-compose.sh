#!/usr/bin/env bash

# Container Orchestration Detection and Selection
# Prefers podman-compose over docker-compose when available and working

COMPOSE_CMD=""
COMPOSE_TYPE=""

# Function to test if a compose command is working
test_compose_command() {
    local cmd="$1"
    local name="$2"
    
    # Test basic functionality
    if command -v "$cmd" >/dev/null 2>&1; then
        # Test if it can show version without errors
        if timeout 5s "$cmd" version >/dev/null 2>&1; then
            echo "✅ $name is available and working"
            return 0
        else
            echo "⚠️ $name is available but not working properly"
            return 1
        fi
    else
        echo "❌ $name is not available"
        return 1
    fi
}

# Function to detect and set the best compose command
detect_compose() {
    echo "🔍 Detecting container orchestration tools..."
    
    # Try podman-compose first (preferred)
    if test_compose_command "podman-compose" "podman-compose"; then
        COMPOSE_CMD="podman-compose"
        COMPOSE_TYPE="podman"
        echo "🐋 Using podman-compose"
        return 0
    fi
    
    # Fallback to docker-compose
    if test_compose_command "docker-compose" "docker-compose"; then
        COMPOSE_CMD="docker-compose"
        COMPOSE_TYPE="docker"
        echo "🐳 Using docker-compose"
        return 0
    fi
    
    # Try docker compose (newer syntax)
    if command -v docker >/dev/null 2>&1; then
        if timeout 5s docker compose version >/dev/null 2>&1; then
            COMPOSE_CMD="docker compose"
            COMPOSE_TYPE="docker"
            echo "🐳 Using docker compose (plugin)"
            return 0
        fi
    fi
    
    echo "❌ No working container orchestration tool found!"
    echo "Please install one of the following:"
    echo "  - podman-compose (preferred)"
    echo "  - docker-compose"
    echo "  - docker with compose plugin"
    return 1
}

# Function to get compose command
get_compose_command() {
    if [ -z "$COMPOSE_CMD" ]; then
        detect_compose >/dev/null
    fi
    echo "$COMPOSE_CMD"
}

# Function to get compose type
get_compose_type() {
    if [ -z "$COMPOSE_TYPE" ]; then
        detect_compose >/dev/null
    fi
    echo "$COMPOSE_TYPE"
}

# Function to run compose command with error handling
run_compose() {
    local cmd
    cmd="$(get_compose_command)"
    
    if [ -z "$cmd" ]; then
        echo "❌ No compose command available" >&2
        return 1
    fi
    
    echo "🔧 Running: $cmd $*" >&2
    $cmd "$@"
}

# Function to check if compose is working with our files
test_compose_config() {
    local compose_file="${1:-docker-compose.yml}"
    local cmd
    cmd="$(get_compose_command)"
    
    if [ -z "$cmd" ]; then
        return 1
    fi
    
    echo "🧪 Testing compose configuration: $compose_file"
    
    # Test config validation
    if $cmd -f "$compose_file" config --quiet 2>/dev/null; then
        echo "✅ Compose configuration is valid"
        return 0
    else
        echo "❌ Compose configuration is invalid"
        return 1
    fi
}

# Function to get appropriate compose file
get_compose_file() {
    local compose_type
    compose_type="$(get_compose_type)"
    
    # For podman, prefer the simple compose file to avoid networking issues
    if [ "$compose_type" = "podman" ]; then
        if [ -f "docker-compose.simple.yml" ]; then
            echo "docker-compose.simple.yml"
            return 0
        fi
    fi
    
    # Default to main compose file
    if [ -f "docker-compose.yml" ]; then
        echo "docker-compose.yml"
    else
        echo "docker-compose.simple.yml"
    fi
}

# Export functions for use in other scripts
export -f get_compose_command
export -f get_compose_type  
export -f run_compose
export -f test_compose_config
export -f get_compose_file

# If script is run directly, show detection results
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    detect_compose
    
    echo
    echo "📋 Detection Results:"
    echo "  Command: $(get_compose_command)"
    echo "  Type: $(get_compose_type)"
    echo "  Suggested compose file: $(get_compose_file)"
    
    # Test configuration if compose files exist
    echo
    echo "🧪 Testing configurations..."
    for file in docker-compose.yml docker-compose.simple.yml; do
        if [ -f "$file" ]; then
            test_compose_config "$file"
        fi
    done
fi