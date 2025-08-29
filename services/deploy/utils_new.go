package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

func (s *service) findFile(filePath string) error {
	_, err := os.Stat(filePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file does not exist")
		}

		return errors.New("unexpected error")
	}
	return err
}

func (s *service) gitClone(repo string, clonePath string) error {
	var stderr, stdout bytes.Buffer

	if err := os.MkdirAll(clonePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", clonePath, err)
	}

	cmd := exec.Command("git", "clone", repo, ".")
	cmd.Dir = clonePath
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git clone failed %v: stderr: %v", err, stderr.String())
	}

	return nil
}

func (s *service) removeDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dirPath, err)
	}
	return nil
}

const charMap = "abcdefghijklmnopqrstuvwxyz" + "0123456789"

func randomStringWithCharMap(length int, charset string) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (s *service) getRandomString(length int) string {
	return randomStringWithCharMap(length, charMap)
}

func (s *service) getClonePath(id string) string {
	return fmt.Sprintf("/projects/%v", id)
}

func (s *service) getCacheKey(id string, userid string) string {
	return fmt.Sprintf("deployment:%s:%s", id, userid)
}
