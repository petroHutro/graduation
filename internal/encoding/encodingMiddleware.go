package encoding

import (
	"encoding/base64"
	"fmt"
)

// func EncodeID(id int) string {
// 	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("id:%d", id)))
// }

// func DecodeID(encodedID string) (int, error) {
// 	decodedBytes, err := base64.StdEncoding.DecodeString(encodedID)
// 	if err != nil {
// 		return 0, err
// 	}

// 	var id int
// 	_, err = fmt.Sscanf(string(decodedBytes), "id:%d", &id)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return id, nil
// }

var secretKey = "mySecretKey"

func EncodeID(id int) string {
	idBytes := []byte(fmt.Sprintf("%d", id))

	encodedBytes := []byte(secretKey)
	for i := range idBytes {
		idBytes[i] ^= encodedBytes[i%len(encodedBytes)]
	}

	return base64.StdEncoding.EncodeToString(idBytes)
}

func DecodeID(encodedID string) (int, error) {
	idBytes, err := base64.StdEncoding.DecodeString(encodedID)
	if err != nil {
		return 0, err
	}

	encodedBytes := []byte(secretKey)
	for i := range idBytes {
		idBytes[i] ^= encodedBytes[i%len(encodedBytes)]
	}

	var decodedID int
	_, err = fmt.Sscanf(string(idBytes), "%d", &decodedID)
	if err != nil {
		return 0, err
	}

	return decodedID, nil
}
