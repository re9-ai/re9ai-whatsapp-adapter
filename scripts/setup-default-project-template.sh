#!/bin/bash

# =============================================================================
# re9.ai Default Project Template Setup Script
# =============================================================================
# This script configures the standard project structure for re9.ai repositories
# including wiki submodule integration, GitHub Copilot instructions, and 
# documentation templates.
#
# Usage: ./scripts/setup-default-project-template.sh [REPO_NAME]
#
# Arguments:
#   REPO_NAME     - Name of the repository (used in documentation, optional)
#
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_NAME="${1:-$(basename "$PROJECT_ROOT")}"
WIKI_REPO_URL="https://github.com/re9-ai/wiki.git"

# =============================================================================
# Helper Functions
# =============================================================================

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

# =============================================================================
# Setup Functions
# =============================================================================

setup_scripts_directory() {
    print_header "Setting up scripts directory"
    
    if [[ ! -d "$PROJECT_ROOT/scripts" ]]; then
        mkdir -p "$PROJECT_ROOT/scripts"
        print_success "Created scripts directory"
    else
        print_info "Scripts directory already exists"
    fi
}

setup_wiki_submodule() {
    print_header "Setting up wiki submodule"
    
    print_info "Using centralized re9.ai wiki: $WIKI_REPO_URL"
    
    cd "$PROJECT_ROOT"
    
    # Remove existing wiki submodule if it exists
    if [[ -f ".gitmodules" ]] && grep -q "path = wiki" .gitmodules; then
        print_info "Removing existing wiki submodule"
        git submodule deinit -f wiki 2>/dev/null || true
        git rm --cached wiki 2>/dev/null || true
        rm -rf wiki .git/modules/wiki
    fi
    
    # Add wiki as submodule
    print_info "Adding wiki submodule: $WIKI_REPO_URL"
    git submodule add -b main "$WIKI_REPO_URL" wiki
    
    # Configure submodule to ignore local changes
    git config -f .gitmodules submodule.wiki.ignore dirty
    
    # Initialize and update submodule
    git submodule init
    git submodule update --remote wiki
    
    print_success "Wiki submodule configured"
}

create_wiki_sync_script() {
    print_header "Creating wiki sync script"
    
    cat > "$PROJECT_ROOT/scripts/update-wiki.sh" << 'EOF'
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
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

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
git submodule update --remote wiki

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
EOF

    chmod +x "$PROJECT_ROOT/scripts/update-wiki.sh"
    print_success "Wiki sync script created and made executable"
}

create_readme_template() {
    print_header "Setting up README.md template"
    
    # Check if README exists and if it already has the wiki section
    if [[ -f "$PROJECT_ROOT/README.md" ]]; then
        if grep -q "## 📚 Documentation & Wiki" "$PROJECT_ROOT/README.md"; then
            print_info "README.md already contains wiki documentation section"
            return
        fi
    fi
    
    # Create or update README with wiki section at the beginning
    local temp_readme=$(mktemp)
    
    # Wiki documentation header
    cat > "$temp_readme" << EOF
## 📚 Documentation & Wiki

This repository includes a comprehensive wiki submodule containing architectural and business documentation:

- **📖 Wiki Access**: The \`wiki/\` directory contains extensive documentation
- **🔄 Wiki Sync**: Use \`./scripts/update-wiki.sh\` to sync latest wiki changes
- **📋 Wiki Guide**: See [WIKI.md](WIKI.md) for detailed wiki management instructions

⚠️ **Important**: The wiki is read-only in this repository. Edit content through GitHub's wiki interface, then sync locally using the provided script.

EOF

    # If README exists, append its content (unless it's just a default GitHub README)
    if [[ -f "$PROJECT_ROOT/README.md" ]]; then
        # Skip if it's just a default README
        if ! grep -q "^# $REPO_NAME$" "$PROJECT_ROOT/README.md" && [[ $(wc -l < "$PROJECT_ROOT/README.md") -gt 5 ]]; then
            echo "" >> "$temp_readme"
            cat "$PROJECT_ROOT/README.md" >> "$temp_readme"
        else
            # Add basic project header
            cat >> "$temp_readme" << EOF
# $REPO_NAME

[Project description goes here]

## Getting Started

[Getting started instructions go here]

## Documentation

For detailed documentation, architecture diagrams, and business requirements, see the [wiki](./wiki/) directory.

EOF
        fi
    else
        # Create basic README structure
        cat >> "$temp_readme" << EOF
# $REPO_NAME

[Project description goes here]

## Getting Started

[Getting started instructions go here]

## Documentation

For detailed documentation, architecture diagrams, and business requirements, see the [wiki](./wiki/) directory.

EOF
    fi
    
    mv "$temp_readme" "$PROJECT_ROOT/README.md"
    print_success "README.md updated with wiki documentation section"
}

create_wiki_md() {
    print_header "Creating WIKI.md instructions"
    
    cat > "$PROJECT_ROOT/WIKI.md" << 'EOF'
# Wiki Submodule Management

The `wiki/` directory is a read-only submodule that mirrors the GitHub wiki repository.

## Important: Wiki Editing Guidelines

### ✅ How to Edit Wiki Content
- **Use GitHub's wiki interface**: Navigate to the wiki tab in GitHub and edit pages there
- **Work directly on wiki repo**: Clone the re9.ai/wiki repository separately and work on it directly
- **Use GitHub web editor**: Edit wiki files directly through the GitHub web interface
- **Collaborate through GitHub**: Use GitHub's built-in wiki collaboration features

### ❌ What NOT to Do
- **Do NOT edit files directly** in the `wiki/` directory of this repository
- **Do NOT commit changes** to files inside the `wiki/` directory
- **Do NOT use git commands** inside the `wiki/` directory

## Syncing Latest Wiki Changes

When wiki content is updated through GitHub's wiki interface, use these commands to sync the changes locally:

### Automatic Sync (Recommended)
```bash
# Use the provided script for safe wiki updates
./scripts/update-wiki.sh
```

The script will:
- Check for local changes and warn you if any exist
- Reset local changes (since wiki should be read-only)
- Update the submodule to the latest remote version
- Commit the submodule reference update automatically

### Manual Sync
```bash
# Pull latest wiki changes manually
git submodule update --remote wiki

# Commit the submodule reference update
git add wiki
git commit -m "docs: sync wiki submodule to latest version"
```

## If You Accidentally Made Local Changes

If you accidentally edited files in the wiki directory, you can reset them:

```bash
# Navigate to wiki directory and reset all changes
cd wiki
git reset --hard HEAD
git clean -fd
cd ..

# Then sync with remote
./scripts/update-wiki.sh
```

## Wiki Structure Guidelines

When working on the wiki repository directly, organize content using this structure:

### Architecture Documentation (`architecture/`)
- System architecture and design documents
- Component diagrams and technical specifications
- Infrastructure and deployment documentation
- Security and authentication specifications

### Business Documentation (`business/`)
- Data models and entity relationships
- User stories and requirements
- Business logic and workflows
- Integration specifications

### Root Level Documentation
- README.md - Wiki overview and navigation
- Project vision and storytelling documents

## Pre-commit Hook Integration

This repository includes pre-commit hooks that will:
- Prevent accidental commits to wiki files
- Check for wiki submodule updates
- Validate documentation links

## Troubleshooting

### Submodule Issues
If you encounter submodule issues:

```bash
# Reinitialize submodule
git submodule deinit -f wiki
git submodule update --init --recursive wiki

# Or completely reset submodule
git rm --cached wiki
git submodule add -b main https://github.com/re9-ai/wiki.git wiki
```

### Permission Issues
If you get permission errors when updating the wiki:
1. Ensure you have write access to the wiki repository
2. Check your GitHub authentication (SSH keys or token)
3. Verify the wiki repository URL in `.gitmodules`

## Best Practices

1. **Always sync before starting work**: Run `./scripts/update-wiki.sh` before working with documentation
2. **Use descriptive commit messages**: When the script commits submodule updates, it uses standardized messages
3. **Regular syncing**: Update the wiki submodule regularly to stay current with documentation changes
4. **Separation of concerns**: Keep infrastructure code and documentation separate but synchronized
EOF

    print_success "WIKI.md instructions created"
}

create_copilot_instructions() {
    print_header "Creating GitHub Copilot instructions"
    
    # Create .github directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/.github"
    
    cat > "$PROJECT_ROOT/.github/copilot-instructions.md" << EOF
# GitHub Copilot Instructions for $REPO_NAME

## Repository Overview

This repository contains **[PROJECT_TYPE]** for the re9.ai platform - a chat-first renovation and budgeting platform. [CUSTOMIZE THIS DESCRIPTION BASED ON PROJECT TYPE]

## Project Context

**re9.ai** is a renovation and budgeting platform that connects users with construction managers, builders, material sellers, and service providers through chat interfaces (WhatsApp, Telegram, web chat). The platform leverages AI to understand user needs and facilitate connections between all stakeholders in renovation projects.

## Repository Structure

### Core Files
- [LIST YOUR MAIN PROJECT FILES HERE]
- README.md - Project overview and setup instructions
- WIKI.md - Wiki submodule management guide

### Automation Scripts
- \`scripts/\` - Project management and automation scripts:
  - \`setup-default-project-template.sh\` - Initial project template setup
  - \`update-wiki.sh\` - Wiki submodule synchronization script
  - [ADD OTHER PROJECT-SPECIFIC SCRIPTS]

## Wiki Submodule Documentation Context

⚠️ **IMPORTANT**: This repository includes a \`wiki/\` submodule that contains comprehensive architectural and business documentation.

### Wiki Structure and Content
The \`wiki/\` directory is a **read-only submodule** that mirrors the GitHub wiki repository. It contains:

#### Architecture Documentation (\`wiki/architecture/\`)
- System architecture and design documents
- Component diagrams and technical specifications
- Infrastructure and deployment documentation
- Security and authentication specifications

#### Business Documentation (\`wiki/business/\`)
- Data model documentation with core entities and relationships
- User stories for all platform personas
- Business requirements and workflow specifications
- Integration and payment processing requirements

#### Root Level Documentation
- Project vision and storytelling documents
- Development guidelines and best practices

### Wiki Usage Guidelines for Copilot
1. **Always reference wiki content** when providing architectural context or business logic explanations
2. **Use wiki documentation** to understand the broader system design when suggesting changes
3. **Reference specific wiki files** when explaining how components support business requirements
4. **Consider chat integration requirements** when working on communication or API features

⚠️ **Wiki Editing Restrictions**: 
- The wiki submodule is **read-only** in this repository
- Wiki content should only be edited through GitHub's wiki interface or by working on the wiki repo directly
- Use \`./scripts/update-wiki.sh\` to sync latest wiki changes locally
- Never commit changes directly to files in the \`wiki/\` directory

## Technology Stack

### Core Technologies
[CUSTOMIZE BASED ON YOUR PROJECT TYPE - EXAMPLES:]
- **For Infrastructure**: Terraform, AWS services, Kubernetes
- **For Backend**: Node.js, TypeScript, PostgreSQL, Redis
- **For Frontend**: React, Next.js, TypeScript, Tailwind CSS
- **For Mobile**: React Native, Expo

### Development Environment
[CUSTOMIZE BASED ON YOUR PROJECT REQUIREMENTS]

## Key Architectural Principles

1. **Chat-First Design** - All components support multiple chat platforms (WhatsApp, Telegram)
2. **AI Integration** - Ready for AI/ML workloads and automation
3. **Microservices Architecture** - Modular, scalable component design
4. **Infrastructure as Code** - Everything defined declaratively
5. **Multi-Environment Support** - Dev, staging, production environments
6. **Security by Design** - Encryption, authentication, authorization
7. **Cost Optimization** - Efficient resource utilization

## Development Guidelines for Copilot

### When Suggesting Changes
1. **Always consider the chat integration requirements** from the wiki
2. **Reference the system architecture** documentation
3. **Ensure changes align with the platform strategy**
4. **Consider the impact on AI/ML integration requirements**
5. **Maintain consistency with project patterns and conventions**

### When Working with Business Logic
1. **Reference relevant user stories** from \`wiki/business/user-stories/\`
2. **Use the data model documentation** to understand entity relationships
3. **Reference chat integration patterns** for communication requirements
4. **Consider payment processing requirements** for financial operations

### When Explaining Technical Concepts
1. **Always reference wiki documentation** for authoritative information
2. **Use specific wiki files** to support technical explanations
3. **Consider the broader platform context** when making recommendations
4. **Reference architecture decisions** documented in the wiki

## Common Development Scenarios

### Setting Up Development Environment
\`\`\`bash
# [ADD PROJECT-SPECIFIC SETUP COMMANDS]
# Example for various project types:

# For Node.js projects:
npm install
npm run dev

# For Python projects:
pip install -r requirements.txt
python manage.py runserver

# For Terraform projects:
terraform init
terraform plan
\`\`\`

### Syncing Documentation
\`\`\`bash
# Always sync wiki before starting work
./scripts/update-wiki.sh

# Check wiki status
git status wiki/
\`\`\`

### [ADD OTHER PROJECT-SPECIFIC SCENARIOS]

## Security and Compliance Notes

- **Data Privacy**: LGPD (Brazilian GDPR) compliance required
- **Authentication**: Professional verification layer
- **Encryption**: At-rest and in-transit encryption mandatory
- **Access Control**: Principle of least privilege
- **Audit Trails**: Comprehensive logging and monitoring

## Integration Points

### Chat Platforms
- **WhatsApp Business API** - Primary communication channel
- **Telegram Bot API** - Secondary communication channel
- **Web Chat** - Browser-based communication

### AI/ML Services
- **Amazon Bedrock** - AI/ML capabilities
- **Natural Language Processing** - Chat understanding
- **Document Processing** - File analysis and extraction

### Payment Processing
- **Brazilian Payment Methods** - PIX, credit cards, bank transfers
- **International Payments** - Credit cards, PayPal
- **Compliance** - PCI DSS, financial regulations

Remember: When in doubt about business requirements or architectural decisions, always reference the comprehensive documentation in the \`wiki/\` submodule!
EOF

    print_success "GitHub Copilot instructions created"
    print_warning "Please customize the placeholders in .github/copilot-instructions.md:"
    print_info "  - [PROJECT_TYPE]: Describe what this repository contains"
    print_info "  - Technology stack section: Add your specific technologies"
    print_info "  - Development commands: Add your project-specific commands"
}

setup_gitignore_and_hooks() {
    print_header "Setting up git configuration"
    
    # Add wiki ignore to .gitignore if it doesn't exist
    if [[ ! -f "$PROJECT_ROOT/.gitignore" ]] || ! grep -q "# Wiki submodule - ignore local changes" "$PROJECT_ROOT/.gitignore"; then
        cat >> "$PROJECT_ROOT/.gitignore" << 'EOF'

# Wiki submodule - ignore local changes
wiki/*
!wiki/.gitkeep
EOF
        print_success "Added wiki ignore rules to .gitignore"
    fi
    
    # Create pre-commit hook to prevent wiki commits
    mkdir -p "$PROJECT_ROOT/.git/hooks"
    cat > "$PROJECT_ROOT/.git/hooks/pre-commit" << 'EOF'
#!/bin/bash

# Pre-commit hook to prevent accidental wiki commits
# Check if any wiki files are staged for commit
if git diff --cached --name-only | grep -q "^wiki/"; then
    echo "❌ Error: Attempting to commit wiki files!"
    echo ""
    echo "The wiki submodule is read-only in this repository."
    echo "Please edit wiki content through GitHub's wiki interface."
    echo ""
    echo "To sync wiki changes: ./scripts/update-wiki.sh"
    echo "To unstage wiki files: git reset HEAD wiki/"
    echo ""
    exit 1
fi
EOF
    chmod +x "$PROJECT_ROOT/.git/hooks/pre-commit"
    print_success "Created pre-commit hook to prevent wiki commits"
}

show_completion_summary() {
    print_header "Setup Complete! 🎉"
    
    echo -e "\n${GREEN}The following components have been configured:${NC}"
    echo -e "  ✅ Scripts directory: ${BLUE}scripts/${NC}"
    echo -e "  ✅ Wiki submodule: ${BLUE}wiki/${NC}"
    echo -e "  ✅ Wiki sync script: ${BLUE}scripts/update-wiki.sh${NC}"
    echo -e "  ✅ README.md with wiki documentation"
    echo -e "  ✅ WIKI.md management guide"
    echo -e "  ✅ GitHub Copilot instructions: ${BLUE}.github/copilot-instructions.md${NC}"
    echo -e "  ✅ Git hooks and ignore rules"
    
    echo -e "\n${BLUE}📋 Next Steps:${NC}"
    echo -e "  1. ${YELLOW}Customize GitHub Copilot instructions${NC} in .github/copilot-instructions.md"
    echo -e "  2. ${YELLOW}Update README.md${NC} with your project-specific content"
    echo -e "  3. ${YELLOW}Sync wiki${NC}: ./scripts/update-wiki.sh"
    echo -e "  4. ${YELLOW}Commit changes${NC}: git add . && git commit -m 'feat: setup default project template'"
    
    echo -e "\n${BLUE}📖 Documentation:${NC}"
    echo -e "  • Wiki management: ${BLUE}WIKI.md${NC}"
    echo -e "  • Project setup: ${BLUE}README.md${NC}"
    echo -e "  • AI assistant context: ${BLUE}.github/copilot-instructions.md${NC}"
    
    echo -e "\n${GREEN}Template setup completed successfully!${NC}"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    print_header "re9.ai Default Project Template Setup"
    print_info "Repository: $REPO_NAME"
    print_info "Location: $PROJECT_ROOT"
    
    # Confirm before proceeding
    echo -e "\n${YELLOW}This script will set up the standard re9.ai project template.${NC}"
    echo -e "${YELLOW}It will create/modify files in your repository.${NC}"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Setup cancelled"
        exit 0
    fi
    
    # Execute setup steps
    setup_scripts_directory
    setup_wiki_submodule
    create_wiki_sync_script
    create_readme_template
    create_wiki_md
    create_copilot_instructions
    setup_gitignore_and_hooks
    
    show_completion_summary
}

# Run main function
main "$@"
