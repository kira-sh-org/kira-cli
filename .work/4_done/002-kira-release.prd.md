---
id: 002
title: Kira Release
status: done
kind: prd
assigned: wkallan1984@gmail.com
estimate: 1 day
created: 2025-10-08T15:00:00Z
---
# Kira Release

This PRD outlines the release of Kira, a git-based, plaintext productivity tool designed with both clankers (LLMs) and meatbags (people) in mind.

## Context

Kira is a git-based, plaintext productivity tool designed with both clankers (LLMs) and meatbags (people) in mind. It uses markdown files, git, and a lightweight CLI to manage and coordinate work.

## Technology Stack

The CLI will be built with Go and Cobra for command-line interface management, ensuring fast performance and cross-platform compatibility.

It should be released as a binary for Linux, macOS, and Windows.

It should use github actions to do the following:
- Run tests
- Run linting
- Run formatting
- Run security scanning
- Run code coverage
- Build the binary for Linux, macOS, and Windows
- Release the binary to the github releases page
- Release the binary to the homebrew tap
- Release the binary to the scoop bucket
- Release the binary to the chocolatey package manager