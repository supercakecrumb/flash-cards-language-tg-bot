package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPrivateBankInvitations(t *testing.T) {
	b := &Bot{inviteSecret: []byte(strings.Repeat("test", 8))}
	code, err := b.createBankInvite(42, "receiver")
	if err != nil {
		t.Fatal(err)
	}
	if id, invited, err := b.parseBankInvite(code, "receiver"); err != nil || id != 42 || !invited {
		t.Fatal("valid invite denied")
	}
	for _, test := range []struct{ code, user string }{{code, "someoneelse"}, {code + "x", "receiver"}, {"forged", "receiver"}} {
		if _, _, err := b.parseBankInvite(test.code, test.user); err == nil {
			t.Fatal("invalid invite accepted")
		}
	}
	if id, invited, err := b.parseBankInvite("42", "receiver"); err != nil || id != 42 || invited {
		t.Fatal("numeric IDs must not grant private access")
	}
	b.inviteSecret = nil
	if _, err := b.createBankInvite(42, "receiver"); err == nil {
		t.Fatal("missing signing key accepted")
	}
	if _, _, err := b.parseBankInvite(code, "receiver"); err == nil {
		t.Fatal("missing verification key accepted")
	}
}

func TestExpiredInvitationRejected(t *testing.T) {
	b := &Bot{inviteSecret: []byte(strings.Repeat("test", 8))}
	data, err := json.Marshal(bankInvite{42, "receiver", time.Now().Add(-time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(data)
	mac := hmac.New(sha256.New, b.inviteSecret)
	mac.Write([]byte(payload))
	code := payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if _, _, err := b.parseBankInvite(code, "receiver"); err == nil {
		t.Fatal("expired invitation accepted")
	}
}
