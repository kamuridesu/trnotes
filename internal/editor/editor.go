package editor

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func SaveFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), os.ModeAppend)
}

func ConvertToHtml(text string) string {
	return strings.ReplaceAll(text, "\n", "</br>")
}

func ReadFile(filename string) (*string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	str := ConvertToHtml(string(data))
	return &str, nil
}

func CreateTempFile() (string, error) {
	file, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	tempFileLocation := file.Name()
	if _, err := os.Stat(tempFileLocation); err == nil {
		if err := SaveFile(tempFileLocation, ""); err != nil {
			return "", err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		os.WriteFile(tempFileLocation, []byte(""), 0644)
	}
	return tempFileLocation, err
}

func OpenEditor() (*string, error) {
	tempFileLocation, err := CreateTempFile()
	defer os.Remove(tempFileLocation)
	if err != nil {
		return nil, err
	}

	command := os.Getenv("EDITOR")
	if command == "" && runtime.GOOS == "windows" {
		command = "notepad"
	} else if command == "" {
		command = "nano"
	}
	cmd := exec.Command(command, tempFileLocation)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	cmd.Wait()

	return ReadFile(tempFileLocation)
}
