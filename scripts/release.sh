#!/bin/bash

set -e

# FixPanic CLI Tool - Automated Release Script
# Usage: ./scripts/release.sh [major|minor|patch]
# Default increment type: patch

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Functions for colored output
print_info() {
    echo -e "${BLUE}ℹ️  ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✅ ${NC}$1"
}

print_warning() {
    echo -e "${YELLOW}⚠️  ${NC}$1"
}

print_error() {
    echo -e "${RED}❌ ${NC}$1"
}

print_step() {
    echo -e "${PURPLE}🚀 ${NC}$1"
}

print_result() {
    echo -e "${CYAN}📋 ${NC}$1"
}

# Parse arguments
INCREMENT_TYPE=${1:-patch}

# Validate increment type
if [[ "$INCREMENT_TYPE" != "major" && "$INCREMENT_TYPE" != "minor" && "$INCREMENT_TYPE" != "patch" ]]; then
    print_error "Invalid increment type '$INCREMENT_TYPE'."
    echo "Usage: $0 [major|minor|patch]"
    echo "Default: patch"
    echo ""
    echo "Examples:"
    echo "  $0           # Increment patch version (v1.1.0 → v1.1.1)"
    echo "  $0 minor     # Increment minor version (v1.1.0 → v1.2.0)"
    echo "  $0 major     # Increment major version (v1.1.0 → v2.0.0)"
    exit 1
fi

echo ""
print_step "Starting automated release process for FixPanic CLI Tool"
print_info "Increment type: $INCREMENT_TYPE"
echo ""

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository. Please run this script from the project root."
    exit 1
fi

# Ensure we're on the main branch
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    print_error "Not on main branch (currently on: $CURRENT_BRANCH)."
    print_info "Please switch to main branch: git checkout main"
    exit 1
fi

print_success "On main branch"

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    print_warning "You have uncommitted changes."
    print_info "Uncommitted files:"
    git status --porcelain
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Release cancelled. Please commit or stash your changes first."
        exit 0
    fi
fi

# Pull latest changes
print_step "Pulling latest changes from origin/main"
if ! git pull origin main; then
    print_error "Failed to pull latest changes from origin/main"
    exit 1
fi

print_success "Latest changes pulled successfully"

# Get the latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
print_result "Latest tag: $LATEST_TAG"

# Validate tag format
if [[ ! $LATEST_TAG =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_warning "Latest tag '$LATEST_TAG' doesn't follow semantic versioning (vX.Y.Z)"
    print_info "Using v0.0.0 as starting point"
    LATEST_TAG="v0.0.0"
fi

# Parse version (remove 'v' prefix)
VERSION=${LATEST_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Validate parsed version components
if [[ ! $MAJOR =~ ^[0-9]+$ ]] || [[ ! $MINOR =~ ^[0-9]+$ ]] || [[ ! $PATCH =~ ^[0-9]+$ ]]; then
    print_error "Failed to parse version components from '$LATEST_TAG'"
    exit 1
fi

# Store original version for display
ORIGINAL_VERSION="$MAJOR.$MINOR.$PATCH"

# If starting from v0.0.0, force major increment for first release
if [[ "$LATEST_TAG" == "v0.0.0" ]]; then
    INCREMENT_TYPE="major"
fi

# Increment version based on type
case $INCREMENT_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="v$MAJOR.$MINOR.$PATCH"

# Display version change
echo ""
print_result "Version change: v$ORIGINAL_VERSION → $NEW_VERSION"

# Check if tag already exists
if git rev-parse "$NEW_VERSION" >/dev/null 2>&1; then
    print_error "Tag '$NEW_VERSION' already exists"
    print_info "Existing tags:"
    git tag --sort=-version:refname | head -5
    exit 1
fi

# Show what will happen
echo ""
print_step "Release Summary"
echo "  Repository: fixpanic/fixpanic-cli-tool"
echo "  Current branch: $CURRENT_BRANCH"
echo "  Previous version: $LATEST_TAG"
echo "  New version: $NEW_VERSION"
echo "  Increment type: $INCREMENT_TYPE"
echo ""

# Confirm before creating tag
print_warning "This will create and push tag: $NEW_VERSION"
print_info "This will trigger the GitHub Actions workflow for building and publishing releases."
echo ""
read -p "Continue with release? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_error "Release cancelled."
    exit 0
fi

echo ""

# Create the tag
print_step "Creating tag: $NEW_VERSION"
if ! git tag "$NEW_VERSION"; then
    print_error "Failed to create tag '$NEW_VERSION'"
    exit 1
fi

print_success "Tag created locally"

# Push the tag
print_step "Pushing tag to origin"
if ! git push origin "$NEW_VERSION"; then
    print_error "Failed to push tag to origin"
    print_warning "Tag was created locally but not pushed. You can:"
    print_info "- Retry: git push origin $NEW_VERSION"
    print_info "- Delete local tag: git tag -d $NEW_VERSION"
    exit 1
fi

print_success "Tag pushed to origin successfully"

# Final success message
echo ""
echo "🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉"
print_success "Release $NEW_VERSION created and pushed successfully!"
echo "🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉"
echo ""
print_info "What happens next:"
echo "  1. GitHub Actions workflow will trigger automatically"
echo "  2. Binaries will be built for all platforms"
echo "  3. Release will be published on GitHub"
echo "  4. Assets will be available for download"
echo ""
print_info "Monitor progress:"
echo "  📦 Actions: https://github.com/fixpanic/fixpanic-cli-tool/actions"
echo "  📋 Releases: https://github.com/fixpanic/fixpanic-cli-tool/releases"
echo "  🏷️  Tags: https://github.com/fixpanic/fixpanic-cli-tool/tags"
echo ""
print_info "Test the new release:"
echo "  curl -fsSL https://install.fixpanic.com/install.sh | bash"
echo ""

# Show recent tags for reference
print_info "Recent tags:"
git tag --sort=-version:refname | head -5 | sed 's/^/  /'

echo ""
print_success "Release process completed! 🚀"
echo ""