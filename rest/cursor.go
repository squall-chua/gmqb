package rest

import (
	"encoding/base64"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// EncodeCursor converts a bson.D sort key into a base64url encoded string.
func EncodeCursor(sortKey bson.D, reverse bool) (string, error) {
	if sortKey == nil {
		return "", nil
	}
	// Wrap the sort key and direction into a single BSON document
	data, err := bson.Marshal(bson.D{
		{Key: "k", Value: sortKey},
		{Key: "r", Value: reverse},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a base64url string back into a bson.D and a reverse flag.
func DecodeCursor(token string) (bson.D, bool, error) {
	if token == "" {
		return nil, false, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode cursor token: %w", err)
	}

	var wrapper struct {
		Key     bson.D `bson:"k"`
		Reverse bool   `bson:"r"`
	}
	if err := bson.Unmarshal(data, &wrapper); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cursor data: %w", err)
	}
	return wrapper.Key, wrapper.Reverse, nil
}
