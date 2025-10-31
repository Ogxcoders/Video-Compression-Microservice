package storage

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type WordPressStorage struct {
	apiURL      string
	username    string
	appPassword string
	client      *http.Client
}

func NewWordPressStorage(apiURL, username, appPassword string) *WordPressStorage {
	return &WordPressStorage{
		apiURL:      apiURL,
		username:    username,
		appPassword: appPassword,
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (w *WordPressStorage) DownloadFile(url, destPath string) error {
	resp, err := w.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (w *WordPressStorage) UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	writer.Close()

	uploadURL := w.apiURL + "/media"
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(w.username + ":" + w.appPassword))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload file: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return fmt.Sprintf("%s/uploads/%s", w.apiURL, filepath.Base(filePath)), nil
}

func (w *WordPressStorage) GetFileSize(url string) (int64, error) {
	resp, err := w.client.Head(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get file size: status code %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}
