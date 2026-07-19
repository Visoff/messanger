package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/Visoff/messanger/pkgs/httperrors"
)

func (c *ChatController) UploadChatAvatar(w http.ResponseWriter, r *http.Request) error {
	chat_id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return httperrors.NewHTTPBadRequestError("File too large or invalid form")
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		return httperrors.NewHTTPBadRequestError("No file provided")
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return httperrors.NewHTTPBadRequestError("Only image files are allowed")
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	uuid, err := c.uploadToFileStorage(fileBytes)
	if err != nil {
		return err
	}

	avatarUrl := c.publicFileStorageUrl + "/" + uuid

	chat, err := c.chatService.UpdateChatAvatar(r.Context(), chat_id, avatarUrl)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chat)
	return nil
}

func (c *ChatController) uploadToFileStorage(data []byte) (string, error) {
	req, err := http.NewRequest("POST", c.fileStorageUrl+"/file", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	apiKey := os.Getenv("FILE_STORAGE_API_KEY")
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("file_storage error: %s", string(body))
	}

	var result struct {
		UUID string `json:"uuid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.UUID, nil
}
