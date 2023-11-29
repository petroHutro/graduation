package encoding

import (
	"encoding/base64"
	"fmt"
)

func EncodeID(id int) string {
	idBytes := []byte(fmt.Sprintf("%d", id))
	return base64.StdEncoding.EncodeToString(idBytes)
}

func DecodeID(encodedID string) (int, error) {
	idBytes, err := base64.StdEncoding.DecodeString(encodedID)
	if err != nil {
		return 0, fmt.Errorf("cannot decode string: %v", err)
	}

	var decodedID int
	_, err = fmt.Sscanf(string(idBytes), "%d", &decodedID)
	if err != nil {
		return 0, fmt.Errorf("cannot sscanf: %v", err)
	}

	return decodedID, nil
}
