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
