package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Initialize Git repository
	runCommand("git", "init")

	// Add all files
	runCommand("git", "add", ".")

	// Commit changes
	runCommand("git", "commit", "-m", "Initial commit: Added API Security Scanner project files")

	// Add remote repository
	runCommand("git", "remote", "add", "origin", "git@github.com:elliotsecops/api-security-scanner.git")

	// Push changes
	runCommand("git", "push", "-u", "origin", "master")

	// Create GitHub Actions workflow
	createGitHubActionsWorkflow()

	// Create a new branch
	runCommand("git", "checkout", "-b", "feature/new-feature")

	// Commit changes to the new branch
	runCommand("git", "add", ".")
	runCommand("git", "commit", "-m", "Add new feature")

	// Push the new branch
	runCommand("git", "push", "origin", "feature/new-feature")

	fmt.Println("Automation completed successfully!")
}

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Command execution failed: %v\n", err)
		os.Exit(1)
	}
}

func createGitHubActionsWorkflow() {
	workflowDir := ".github/workflows"
	workflowFile := "ci.yml"

	// Create the directory if it doesn't exist
	err := os.MkdirAll(workflowDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Create the workflow file
	file, err := os.Create(workflowDir + "/" + workflowFile)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Write the workflow content
	workflowContent := `name: CI

on:
  push:
    branches:
      - master
  pull_request:
    branches:
      - master

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v2

    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.17

    - name: Run tests
      run: go test ./...
`
	_, err = file.WriteString(workflowContent)
	if err != nil {
		fmt.Printf("Failed to write to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("GitHub Actions workflow created successfully!")
}