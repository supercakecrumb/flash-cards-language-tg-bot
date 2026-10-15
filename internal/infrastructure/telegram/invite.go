package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type bankInvite struct {
	BankID   int    `json:"bank"`
	Username string `json:"user"`
	Expires  int64  `json:"exp"`
}

var inviteUsername = regexp.MustCompile(`^[a-z0-9_]{5,32}$`)

func (b *Bot) createBankInvite(bankID int, username string) (string, error) {
	username = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(username), "@"))
	if len(b.inviteSecret) < 32 || !inviteUsername.MatchString(username) || bankID <= 0 {
		return "", errors.New("invalid invitation configuration or recipient")
	}
	data, err := json.Marshal(bankInvite{bankID, username, time.Now().Add(24 * time.Hour).Unix()})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(data)
	mac := hmac.New(sha256.New, b.inviteSecret)
	mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (b *Bot) parseBankInvite(code, username string) (int, bool, error) {
	// Numeric IDs remain useful for intentionally public banks, never private ones.
	if id, err := strconv.Atoi(code); err == nil && id > 0 {
		return id, false, nil
	}
	invalid := errors.New("invalid invitation")
	if len(b.inviteSecret) < 32 || len(code) > 1024 {
		return 0, false, invalid
	}
	parts := strings.Split(code, ".")
	if len(parts) != 2 {
		return 0, false, invalid
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, false, invalid
	}
	mac := hmac.New(sha256.New, b.inviteSecret)
	mac.Write([]byte(parts[0]))
	if !hmac.Equal(mac.Sum(nil), sig) {
		return 0, false, invalid
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, false, invalid
	}
	var invite bankInvite
	if json.Unmarshal(data, &invite) != nil || invite.BankID <= 0 || invite.Expires <= time.Now().Unix() || invite.Username != strings.ToLower(username) {
		return 0, false, invalid
	}
	return invite.BankID, true, nil
}
