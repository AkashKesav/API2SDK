package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"

	"github.com/AkashKesav/API2SDK/internal/crypto"
)

// EncryptedString is a string that is automatically encrypted when stored in MongoDB.
type EncryptedString string

// MarshalBSONValue implements the bson.ValueMarshaler interface.
func (es EncryptedString) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if es == "" {
		return bsontype.String, []byte(""), nil
	}
	encrypted, err := crypto.Encrypt(string(es))
	if err != nil {
		return bsontype.Null, nil, err
	}
	return bson.MarshalValue(encrypted)
}

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface.
func (es *EncryptedString) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.Null {
		*es = ""
		return nil
	}
	var encrypted string
	if err := bson.Unmarshal(data, &encrypted); err != nil {
		// Attempt to handle raw string values gracefully.
		if t == bsontype.String {
			*es = EncryptedString(string(data))
			return nil
		}
		return err
	}
	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		// If decryption fails, it might be a raw (unencrypted) value.
		// For backward compatibility, we can choose to return the raw value.
		*es = EncryptedString(encrypted)
		return nil
	}
	*es = EncryptedString(decrypted)
	return nil
}
