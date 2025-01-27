package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func GetTextEditor() string {
	command := os.Getenv("EDITOR")
	if command == "" {
		command = "nano"
		if runtime.GOOS == "windows" {
			command = "notepad"
		}
	}
	return command
}

func SaveFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), os.ModeAppend)
}

func ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func CreateTempFile(content string) (string, error) {
	file, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	tempFileLocation := file.Name()
	if runtime.GOOS == "windows" {
		tempFileLocation += ".txt"
	}
	if _, err := os.Stat(tempFileLocation); err == nil {
		if err := SaveFile(tempFileLocation, content); err != nil {
			return "", err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		os.WriteFile(tempFileLocation, []byte(content), 0644)
	}
	return tempFileLocation, err
}

func EditFile(content string) (string, error) {
	tempFileLocation, err := CreateTempFile(content)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file, error is '%s'", err)
	}
	defer os.Remove(tempFileLocation)
	return OpenEditor(tempFileLocation)
}

func NewFile() (string, error) {
	tempFileLocation, err := CreateTempFile("")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file, error is '%s'", err)
	}
	defer os.Remove(tempFileLocation)
	return OpenEditor(tempFileLocation)
}

func OpenEditor(tempFileLocation string) (string, error) {
	command := GetTextEditor()

	cmd := exec.Command(command, tempFileLocation)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error opening %s, error is '%s'", command, err)
	}
	err := cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("%s failed to execute, error is '%s'", command, err)
	}

	return ReadFile(tempFileLocation)
}
