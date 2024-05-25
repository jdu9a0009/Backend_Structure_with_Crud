package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	AvatarSize int64 = 31457280
)

// Upload returns file_url, file_original_name, file_type, error
func Upload(ctx context.Context, file *multipart.FileHeader, folder string, maxFileSize ...int64) (string, string, string, error) {
	if len(maxFileSize) > 0 {
		if file.Size > maxFileSize[0] {
			return "", "", "", errors.New("file size is too large")
		}
	} else {
		if file.Size > 31457280 {
			return "", "", "", errors.New("file size is too large")
		}
	}
	today := time.Now().Format("02-01-2006")
	initialFolderUrl := fmt.Sprintf("/media/%s/", today)

	if file == nil {
		return "", "", "", nil
	}

	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321"

	i := 0
	randName := ""
	for i < 30 {
		var randSpeed = rand.New(rand.NewSource(time.Now().UnixNano()))
		randIdx := randSpeed.Intn(62)

		randName += string(chars[randIdx])

		i++
	}

	contentTypes := map[string]map[string]interface{}{
		"application/msword": map[string]interface{}{
			"type":       "",
			"permission": true,
		},
		"image/jpeg": map[string]interface{}{
			"type":       "",
			"permission": true,
		},
		"image/png": map[string]interface{}{
			"type":       "",
			"permission": true,
		},
		"application/pdf": map[string]interface{}{
			"type":       "",
			"permission": true,
		},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": map[string]interface{}{
			"type":       "",
			"permission": true,
		},
	}

	if (len(file.Header.Values("Content-Type")) > 0) && (contentTypes[file.Header.Values("Content-Type")[0]] != nil) && (!contentTypes[file.Header.Values("Content-Type")[0]]["permission"].(bool)) {
		return "", "", "", errors.New("content-type of this file has not permission to upload into the server!")
	}

	splitFileName := strings.Split(file.Filename, ".")

	filename := filepath.Base(randName + "." + splitFileName[len(splitFileName)-1])

	if _, err := os.Stat("." + initialFolderUrl + folder); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll("."+initialFolderUrl+folder, os.ModePerm)
		if err != nil {
			return "", "", "", err
		}
	}

	files, err := os.ReadDir("." + initialFolderUrl + folder)

	if err != nil {
		return "", "", "", err
	}

	for _, f := range files {
		if !f.IsDir() && (f.Name() == filename) {
			splitString := strings.Split(filename, ".")
			extra := strconv.Itoa(int(time.Now().Unix()))
			splitString[len(splitString)-2] = splitString[len(splitString)-2] + "-" + extra
			filename = strings.Join(splitString, ".")
			break
		}
	}

	dst := "." + initialFolderUrl + folder + "/" + filename

	src, err := file.Open()
	if err != nil {
		return "", "", "", err
	}
	defer log.Println("file upload src.Close() error: ", src.Close())

	out, err := os.Create(dst)
	if err != nil {
		return "", "", "", err
	}
	//defer log.Println("file upload out.Close() error: ", out.Close())

	_, err = io.Copy(out, src)

	if err != nil {
		return "", "", "", err
	}

	return initialFolderUrl + folder + "/" + filename, filepath.Base(file.Filename), contentTypes[file.Header.Values("Content-Type")[0]]["type"].(string), nil
}

// RemoveFile deletes file in current url
func RemoveFile(ctx context.Context, url string) error {
	err := os.Remove("." + url)

	return err
}

func CheckFileType(ctx context.Context, file *multipart.FileHeader, requiredFileType string) bool {
	contentTypes := map[string]map[string]string{
		"application/msword": map[string]string{
			"type": "docx",
		},
		"image/jpeg": map[string]string{
			"type": "image",
		},
		"image/jpg": map[string]string{
			"type": "image",
		},
		"image/png": map[string]string{
			"type": "image",
		},
		"application/pdf": map[string]string{
			"type": "pdf",
		},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": map[string]string{
			"type": "docx",
		},
	}
	fmt.Println(contentTypes[file.Header.Values("Content-Type")[0]])

	return contentTypes[file.Header.Values("Content-Type")[0]]["type"] == requiredFileType
}
