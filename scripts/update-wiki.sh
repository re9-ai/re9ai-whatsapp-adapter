#!/bin/bash

# =============================================================================
# Wiki Submodule Update Script
# =============================================================================
# This script safely updates the wiki submodule with the latest changes
# from the remote wiki repository.
#
# Usage: ./scripts/update-wiki.sh
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory and project root
SCRIPT_DIR="$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)"
PROJECT_ROOT="$(dirname \"$SCRIPT_DIR\")"

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Helper to update submodule using GITHUB_PAT if available
update_submodule_with_pat() {
    local wiki_url
    wiki_url=$(git config --file .gitmodules submodule.wiki.url)
    if [[ -z "$GITHUB_PAT" ]]; then
        print_error "GITHUB_PAT not set. Cannot update wiki submodule with PAT."
        exit 1
    fi
    # Rewrite URL to use PAT
    local pat_url
    pat_url="https://${GITHUB_PAT}@${wiki_url#https://}"
    print_info "Reconfiguring wiki submodule to use GITHUB_PAT..."
    git config submodule.wiki.url "$pat_url"
    git submodule sync wiki
    git submodule update --remote wiki
}

# Check if wiki submodule exists
if [[ ! -d "$PROJECT_ROOT/wiki" ]]; then
    print_error "Wiki submodule not found. Please run setup-default-project-template.sh first."
    exit 1
fi

print_header "Updating Wiki Submodule"

cd "$PROJECT_ROOT"

# Check for local changes in wiki directory
if [[ -n "$(git status --porcelain wiki)" ]]; then
    print_warning "Local changes detected in wiki directory"
    echo -e "${YELLOW}The wiki submodule should be read-only. Local changes will be reset.${NC}"
    read -p "Continue and reset local changes? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled"
        exit 0
    fi
    
    # Reset local changes
    cd wiki
    git reset --hard HEAD
    git clean -fd
    cd ..
    print_success "Local changes reset"
fi

# Update submodule to latest remote version
print_info "Fetching latest wiki changes..."
if ! git submodule update --remote wiki; then
    print_warning "Wiki submodule update failed. Attempting with GITHUB_PAT if available..."
    update_submodule_with_pat
fi

# Check if there are updates to commit
if [[ -n "$(git status --porcelain wiki)" ]]; then
    print_info "New wiki changes found, updating submodule reference..."
    git add wiki
    git commit -m "docs: sync wiki submodule to latest version

- Updated wiki submodule to latest remote changes
- Automated sync via update-wiki.sh script"
    print_success "Wiki submodule updated and committed"
else
    print_info "Wiki submodule is already up to date"
fi

print_success "Wiki sync completed successfully"

# Display quick summary
echo -e "\n${BLUE}📋 Wiki Summary:${NC}"
echo -e "  • Wiki Location: ${PROJECT_ROOT}/wiki"
echo -e "  • Edit Method: Use GitHub wiki interface or work on the wiki repo directly"
echo -e "  • Sync Command: ./scripts/update-wiki.sh"
echo -e "  • Documentation: See WIKI.md for detailed instructions"
