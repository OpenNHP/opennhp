#!/bin/bash

# OpenNHP Docker Quick Start Script
# Supports Linux and macOS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Default values
USE_CHINA_MIRROR=false
GOPROXY_DEFAULT="https://proxy.golang.org,direct"
GOPROXY_CHINA="https://goproxy.cn,direct"
APT_MIRROR_CHINA="mirrors.aliyun.com"

# Export environment variables
export_env() {
    if [ "$USE_CHINA_MIRROR" = true ]; then
        export GOPROXY="$GOPROXY_CHINA"
        export APT_MIRROR="$APT_MIRROR_CHINA"
        echo -e "${GREEN}Using China mirrors: GOPROXY=$GOPROXY, APT_MIRROR=$APT_MIRROR${NC}"
    else
        export GOPROXY="$GOPROXY_DEFAULT"
        export APT_MIRROR=""
        echo -e "${GREEN}Using default mirrors: GOPROXY=$GOPROXY${NC}"
    fi
}

# Print header
print_header() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║           OpenNHP Docker Quick Start Script                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Print menu
print_menu() {
    echo -e "${YELLOW}Please select an option:${NC}"
    echo ""
    echo "  [1] Build ALL and Start (Full rebuild)"
    echo "  [2] Build Base Image (opennhp-base)"
    echo "  [3] Build NHP-Server"
    echo "  [4] Build NHP-AC"
    echo "  [5] Build NHP-Agent"
    echo "  [6] Build Web-App"
    echo "  [7] Start All Services"
    echo "  [8] Stop All Services"
    echo "  [9] Restart All Services"
    echo "  [10] View Logs (nhp-server)"
    echo "  [11] View Logs (nhp-ac)"
    echo "  [12] View Logs (nhp-agent)"
    echo "  [13] Clean Docker Images"
    echo "  [14] Clean ALL (images + volumes + networks)"
    echo "  [15] Toggle China Mirror (current: $([ "$USE_CHINA_MIRROR" = true ] && echo "ON" || echo "OFF"))"
    echo "  [0] Exit"
    echo ""
}

# Build base image
build_base() {
    echo -e "${BLUE}Building opennhp-base image...${NC}"
    export_env

    local build_args=(--no-cache -t opennhp-base:latest -f Dockerfile.base)
    if [ "$USE_CHINA_MIRROR" = true ]; then
        build_args+=(--build-arg "GOPROXY=$GOPROXY" --build-arg "APT_MIRROR=$APT_MIRROR")
    fi

    docker build "${build_args[@]}" ..
    echo -e "${GREEN}Base image built successfully!${NC}"
}

# Build a specific service
build_service() {
    local service=$1
    echo -e "${BLUE}Building $service...${NC}"
    export_env

    docker compose build --no-cache "$service"
    echo -e "${GREEN}$service built successfully!${NC}"
}

# Build all and start
build_all_and_start() {
    echo -e "${BLUE}Building all images and starting services...${NC}"
    export_env

    # Build base image first
    build_base

    # Build all services
    echo -e "${BLUE}Building all services...${NC}"
    docker compose build --no-cache

    # Stop and remove existing containers
    echo -e "${BLUE}Stopping existing services...${NC}"
    docker compose down 2>/dev/null || true

    # Start services
    echo -e "${BLUE}Starting all services...${NC}"
    docker compose up -d

    echo -e "${GREEN}All services are running!${NC}"
    docker compose ps
}

# Start services
start_services() {
    echo -e "${BLUE}Starting all services...${NC}"
    docker compose up -d
    echo -e "${GREEN}Services started!${NC}"
    docker compose ps
}

# Stop services
stop_services() {
    echo -e "${BLUE}Stopping all services...${NC}"
    docker compose down
    echo -e "${GREEN}Services stopped!${NC}"
}

# Restart services
restart_services() {
    echo -e "${BLUE}Restarting all services...${NC}"
    docker compose restart
    echo -e "${GREEN}Services restarted!${NC}"
    docker compose ps
}

# View logs
view_logs() {
    local service=$1
    echo -e "${BLUE}Viewing logs for $service (Ctrl+C to exit)...${NC}"
    docker compose logs -f "$service"
}

# Clean images
clean_images() {
    echo -e "${YELLOW}This will remove all OpenNHP Docker images.${NC}"
    read -p "Are you sure? (y/N): " confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        echo -e "${BLUE}Stopping services...${NC}"
        docker compose down 2>/dev/null || true

        echo -e "${BLUE}Removing images...${NC}"
        docker rmi opennhp-base:latest 2>/dev/null || true
        docker rmi opennhp-server:latest 2>/dev/null || true
        docker rmi opennhp-ac:latest 2>/dev/null || true
        docker rmi opennhp-agent:latest 2>/dev/null || true
        docker rmi web-app:latest 2>/dev/null || true

        # Also remove by compose project name
        docker images | grep -E "^(opennhp|docker)" | awk '{print $3}' | xargs -r docker rmi 2>/dev/null || true

        echo -e "${GREEN}Images cleaned!${NC}"
    else
        echo -e "${YELLOW}Operation cancelled.${NC}"
    fi
}

# Clean all
clean_all() {
    echo -e "${RED}WARNING: This will remove ALL OpenNHP Docker images, volumes, and networks!${NC}"
    read -p "Are you sure? (y/N): " confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        echo -e "${BLUE}Stopping services...${NC}"
        docker compose down -v 2>/dev/null || true

        echo -e "${BLUE}Removing images...${NC}"
        docker rmi opennhp-base:latest 2>/dev/null || true
        docker rmi opennhp-server:latest 2>/dev/null || true
        docker rmi opennhp-ac:latest 2>/dev/null || true
        docker rmi opennhp-agent:latest 2>/dev/null || true
        docker rmi web-app:latest 2>/dev/null || true
        docker images | grep -E "^(opennhp|docker)" | awk '{print $3}' | xargs -r docker rmi 2>/dev/null || true

        echo -e "${BLUE}Removing volumes...${NC}"
        docker volume ls | grep -E "docker_" | awk '{print $2}' | xargs -r docker volume rm 2>/dev/null || true

        echo -e "${BLUE}Removing networks...${NC}"
        docker network ls | grep -E "docker_" | awk '{print $2}' | xargs -r docker network rm 2>/dev/null || true

        echo -e "${BLUE}Pruning unused Docker resources...${NC}"
        docker system prune -f

        echo -e "${GREEN}All cleaned!${NC}"
    else
        echo -e "${YELLOW}Operation cancelled.${NC}"
    fi
}

# Toggle China mirror
toggle_china_mirror() {
    if [ "$USE_CHINA_MIRROR" = true ]; then
        USE_CHINA_MIRROR=false
        echo -e "${GREEN}China mirror: OFF${NC}"
    else
        USE_CHINA_MIRROR=true
        echo -e "${GREEN}China mirror: ON${NC}"
    fi
}

# Rebuild and restart a specific service
rebuild_and_restart_service() {
    local service=$1
    echo -e "${BLUE}Rebuilding and restarting $service...${NC}"
    export_env

    docker compose build --no-cache "$service"
    docker stop "$service" 2>/dev/null || true
    docker rm "$service" 2>/dev/null || true
    docker compose up -d "$service"

    echo -e "${GREEN}$service rebuilt and restarted!${NC}"
}

# Check Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Error: Docker is not installed or not in PATH${NC}"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        echo -e "${RED}Error: Docker daemon is not running${NC}"
        exit 1
    fi

    if ! command -v docker compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}Error: Docker Compose is not available${NC}"
        exit 1
    fi
}

# Main function
main() {
    check_docker

    # Check for --china flag
    if [[ "$1" == "--china" ]] || [[ "$1" == "-c" ]]; then
        USE_CHINA_MIRROR=true
    fi

    while true; do
        print_header
        print_menu

        read -p "Enter your choice [0-15]: " choice
        echo ""

        case $choice in
            1)
                build_all_and_start
                ;;
            2)
                build_base
                ;;
            3)
                rebuild_and_restart_service "nhp-server"
                ;;
            4)
                rebuild_and_restart_service "nhp-ac"
                ;;
            5)
                rebuild_and_restart_service "nhp-agent"
                ;;
            6)
                rebuild_and_restart_service "web-app"
                ;;
            7)
                start_services
                ;;
            8)
                stop_services
                ;;
            9)
                restart_services
                ;;
            10)
                view_logs "nhp-server"
                ;;
            11)
                view_logs "nhp-ac"
                ;;
            12)
                view_logs "nhp-agent"
                ;;
            13)
                clean_images
                ;;
            14)
                clean_all
                ;;
            15)
                toggle_china_mirror
                ;;
            0)
                echo -e "${GREEN}Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid option. Please try again.${NC}"
                ;;
        esac

        echo ""
        read -p "Press Enter to continue..."
        clear
    done
}

# Run main
main "$@"
