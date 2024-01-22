package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func executeInDocker(goFilePath string) (string, error) {
	// Get the absolute path to the file
	absPath, err := filepath.Abs(goFilePath)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/app/usercode.go", absPath), "golang:latest", "go", "test", "/app/usercode.go")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
