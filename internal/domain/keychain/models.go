package keychain

import (
	"github.com/google/uuid"
	"time"
)

// KeyType represents the type of a key stored in the system.
type KeyType string

const (
	// Need to be added to migration, when add new key. Storage checks contains type field

	// KeyCredential represents a username/password credential.
	KeyCredential KeyType = "credential"
	// KeyText represents arbitrary text data.
	KeyText KeyType = "text"
	// KeyFile represents a file stored as bytes.
	KeyFile KeyType = "file"
	// KeyBankCard represents a bank card data.
	KeyBankCard KeyType = "card"
)

// AllKeyTypes contains all available key types.
var AllKeyTypes = []KeyType{
	KeyCredential,
	KeyText,
	KeyFile,
	KeyBankCard,
}

// String returns the string representation of KeyType.
func (kt KeyType) String() string { return string(kt) }

// CredentialData stores the data for a credential key.
type CredentialData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Site     string `json:"site,omitempty"`
	Note     string `json:"note,omitempty"`
}

// CardData stores the data for a bank card key.
type CardData struct {
	Number  string `json:"number"`
	ExpDate string `json:"exp_date"`
	CVV     string `json:"cvv"`
	Holder  string `json:"holder"`
	Bank    string `json:"bank,omitempty"`
	Note    string `json:"note,omitempty"`
}

// TextData stores arbitrary text data for a key.
type TextData struct {
	Text string `json:"text"`
	Note string `json:"note,omitempty"`
}

// FileData stores a file as byte slice for a key.
type FileData struct {
	File []byte `json:"-"`
	Note string `json:"note,omitempty"`
}

// KeyRecord represents a single key entry in the storage.
type KeyRecord struct {
	ID        int64
	KeyUUID   uuid.UUID
	UserID    int64
	KeyType   KeyType
	Title     string
	Data      []byte
	Nonce     []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ParseKeyType converts a string to a KeyType.
//
// Returns the KeyType and true if the string is valid, or empty string and false otherwise.
func ParseKeyType(s string) (KeyType, bool) {
	switch KeyType(s) {
	case KeyCredential, KeyText, KeyFile, KeyBankCard:
		return KeyType(s), true
	default:
		return "", false
	}
}
