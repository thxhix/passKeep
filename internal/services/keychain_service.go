package services

import (
	"context"
	"encoding/json"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
)

type CryptManager interface {
	Encrypt(plaintext []byte) (nonce []byte, ciphertext []byte, err error)
	Decrypt(nonce []byte, ciphertext []byte) ([]byte, error)
}

type IKeychainService interface {
	GetKeys(ctx context.Context, userID int64, keyType *keychain.KeyType) (list []*keychain.KeyRecord, err error)
	GetKey(ctx context.Context, userID int64, keyUUID string) (record *keychain.KeyRecord, decryptedData []byte, err error)
	DeleteKey(ctx context.Context, userID int64, keyUUID string) error
	AddCredential(ctx context.Context, userID int64, in dto.AddCredentialsDTO) (string, error)
	AddCard(ctx context.Context, userID int64, in dto.AddCardDTO) (string, error)
	AddText(ctx context.Context, userID int64, in dto.AddTextDTO) (string, error)
	AddFile(ctx context.Context, userID int64, in dto.AddFileDTO) (string, error)
}

type KeychainService struct {
	keychainRepo keychain.KeychainRepository
	cryptManager CryptManager
}

func NewKeychainService(keychainRepo keychain.KeychainRepository, cManager CryptManager) KeychainService {
	return KeychainService{
		keychainRepo: keychainRepo,
		cryptManager: cManager,
	}
}

func (s *KeychainService) GetKeys(ctx context.Context, userID int64, keyType *keychain.KeyType) (list []*keychain.KeyRecord, err error) {
	var typ *string
	if keyType != nil {
		t := string(*keyType)
		typ = &t
	}
	return s.keychainRepo.GetUserKeys(ctx, userID, typ)
}

func (s *KeychainService) GetKey(ctx context.Context, userID int64, keyUUID string) (record *keychain.KeyRecord, decryptedData []byte, err error) {
	keyRecord, err := s.keychainRepo.GetUserKey(ctx, userID, keyUUID)
	if err != nil {
		return nil, nil, err
	}

	decryptedData, err = s.cryptManager.Decrypt(keyRecord.Nonce, keyRecord.Data)
	if err != nil {
		return nil, nil, err
	}

	return keyRecord, decryptedData, nil
}

func (s *KeychainService) DeleteKey(ctx context.Context, userID int64, keyUUID string) error {
	return s.keychainRepo.DeleteKey(ctx, userID, keyUUID)
}

func (s *KeychainService) AddCredential(ctx context.Context, userID int64, in dto.AddCredentialsDTO) (string, error) {
	if err := keychain.ValidateTitle(in.Title); err != nil {
		return "", err
	}
	if err := keychain.ValidateCredential(in.Login); err != nil {
		return "", err
	}

	data := keychain.CredentialData{
		Login:    in.Login,
		Password: in.Password,
		Site:     in.Site,
		Note:     in.Note,
	}

	plain, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	nonce, ct, err := s.cryptManager.Encrypt(plain)
	if err != nil {
		return "", err
	}

	uuid, err := s.keychainRepo.AddKey(ctx, userID, keychain.KeyCredential, in.Title, ct, nonce)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (s *KeychainService) AddCard(ctx context.Context, userID int64, in dto.AddCardDTO) (string, error) {
	if err := keychain.ValidateTitle(in.Title); err != nil {
		return "", err
	}
	if err := keychain.ValidateCard(in.Number, in.CVV); err != nil {
		return "", err
	}

	data := keychain.CardData{
		Number:  in.Number,
		ExpDate: in.ExpDate,
		CVV:     in.CVV,
		Holder:  in.Holder,
		Bank:    in.Bank,
		Note:    in.Note,
	}

	plain, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	nonce, ct, err := s.cryptManager.Encrypt(plain)
	if err != nil {
		return "", err
	}

	uuid, err := s.keychainRepo.AddKey(ctx, userID, keychain.KeyBankCard, in.Title, ct, nonce)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (s *KeychainService) AddText(ctx context.Context, userID int64, in dto.AddTextDTO) (string, error) {
	if err := keychain.ValidateTitle(in.Title); err != nil {
		return "", err
	}
	if err := keychain.ValidateText(in.Text); err != nil {
		return "", err
	}

	data := keychain.TextData{
		Text: in.Text,
		Note: in.Note,
	}

	plain, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	nonce, ct, err := s.cryptManager.Encrypt(plain)
	if err != nil {
		return "", err
	}

	uuid, err := s.keychainRepo.AddKey(ctx, userID, keychain.KeyText, in.Title, ct, nonce)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (s *KeychainService) AddFile(ctx context.Context, userID int64, in dto.AddFileDTO) (string, error) {
	if err := keychain.ValidateTitle(in.Title); err != nil {
		return "", err
	}

	data := keychain.FileData{
		File: in.File,
		Note: in.Note,
	}

	plain, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	nonce, ct, err := s.cryptManager.Encrypt(plain)
	if err != nil {
		return "", err
	}

	uuid, err := s.keychainRepo.AddKey(ctx, userID, keychain.KeyFile, in.Title, ct, nonce)
	if err != nil {
		return "", err
	}

	return uuid, nil
}
