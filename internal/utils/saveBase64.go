package utils

import (
	"encoding/base64"
	"os"
	"path/filepath"
)

func SaveBase64Image(filename, base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	err = os.MkdirAll("images", os.ModePerm)
	if err != nil {
		return err
	}

	filePath := filepath.Join("images", filename)
	err = os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
