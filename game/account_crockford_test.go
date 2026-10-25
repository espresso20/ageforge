package game

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

// crockfordGolden pins the recovery-code encoding byte for byte. The pairs were produced by
// the hand-written encoder crockfordEnc replaced (2,820 random payloads matched it before it
// went), so a saved code a player wrote down keeps decoding to the same account.
var crockfordGolden = []struct{ hex, code string }{
	{"", ""},
	{"00", "00"},
	{"ff", "ZW"},
	{"0102030405", "04106105"},
	{"000000000000000000000000000000000000", "00000000000000000000000000000"},
	{"ffffffffffffffffffffffffffffffffffff", "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZY"},
	{"000102030405060708090a0b0c0d0e0f1011", "000G40R40M30E209185GR38E1W812"},
	{"75ed184249e9bc19675e4d1f766213da71b6", "EQPHGGJ9X6Y1JSTY9MFQCRGKV9RVC"},
	{"4278fed2cad5f18a247619205e302653a934", "89WFXMPATQRRM93P34G5WC16AEMK8"},
}

// TestCrockfordGolden covers the encoder against the golden table and that every code
// decodes back to its payload.
func TestCrockfordGolden(t *testing.T) {
	for _, g := range crockfordGolden {
		payload, _ := hex.DecodeString(g.hex)
		if got := crockfordEnc.EncodeToString(payload); got != g.code {
			t.Errorf("encode %s = %q, want %q", g.hex, got, g.code)
		}
		got, err := crockfordDecode(g.code, len(payload))
		if err != nil {
			t.Errorf("decode %q: %v", g.code, err)
		} else if !bytes.Equal(got, payload) {
			t.Errorf("decode %q = %x, want %x", g.code, got, payload)
		}
	}
}

// TestRecoveryCodeGolden covers the whole code (prefix, checksum, dash groups) for four
// fixed account ids, and that each code reads back to its id.
func TestRecoveryCodeGolden(t *testing.T) {
	for id, want := range map[string]string{
		"00000000000000000000000000000000": "AGEF-0000-0000-0000-0000-0000-0000-01N0-M",
		"000102030405060708090a0b0c0d0e0f": "AGEF-000G-40R4-0M30-E209-185G-R38E-1WXK-E",
		"ffffffffffffffffffffffffffffffff": "AGEF-ZZZZ-ZZZZ-ZZZZ-ZZZZ-ZZZZ-ZZZZ-ZXN4-P",
		"9f3c1a7e5b2d4086a1c3e5f7092b4d6e": "AGEF-KWY1-MZJV-5N08-D8E3-WQVG-JATD-DR3K-4",
	} {
		a := &Account{AccountID: id}
		if got := a.RecoveryCode(); got != want {
			t.Errorf("RecoveryCode(%s) = %q, want %q", id, got, want)
		}
		if got, err := RecoveryCodeID(want); err != nil || got != id {
			t.Errorf("RecoveryCodeID(%q) = %q, %v; want %q", want, got, err, id)
		}
	}
}

// TestCrockfordDecodeInput covers the decoder's leniency and its two errors: the
// look-alike letters fold (O to 0, I and L to 1, U to V), a bad symbol names itself, a short
// code says so, and a code longer than the 29 symbols the payload needs is cut to 29 once
// every symbol has been checked.
func TestCrockfordDecodeInput(t *testing.T) {
	const code = "000G40R40M30E209185GR38E1WXKE" // the 29 symbols of the 18-byte payload
	want, err := crockfordDecode(code, 18)
	if err != nil || len(want) != 18 {
		t.Fatalf("decode %q = %x, %v", code, want, err)
	}
	same := func(name, in string) {
		t.Helper()
		got, err := crockfordDecode(in, 18)
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("%s: decode %q = %x, %v; want %x", name, in, got, err, want)
		}
	}
	same("lowercase", strings.ToLower(code))
	same("30 symbols", code+"Z")
	same("40 symbols", code+"ZZZZZZZZZZZ")
	same("O for 0", strings.ReplaceAll(code, "0", "O"))
	same("I for 1", strings.ReplaceAll(code, "1", "I"))
	same("l for 1", strings.ReplaceAll(strings.ToLower(code), "1", "l"))
	if a, err := crockfordDecode("VV", 1); err != nil {
		t.Fatalf("decode VV: %v", err)
	} else if b, err := crockfordDecode("UU", 1); err != nil || !bytes.Equal(a, b) {
		t.Errorf("U should read as V: decode UU = %x, %v; VV = %x", b, err, a)
	}

	for _, c := range []struct{ name, in, msg string }{
		{"bad symbol", "000G40R4*M30E209185GR38E1WXK", "has a character it cannot contain ('*')"},
		{"bad symbol past the end", code + "!", "has a character it cannot contain ('!')"},
		{"short", code[:28], "is too short"},
		{"empty", "", "is too short"},
	} {
		if _, err := crockfordDecode(c.in, 18); err == nil || !strings.Contains(err.Error(), c.msg) {
			t.Errorf("%s: decode %q err = %v, want it to say %q", c.name, c.in, err, c.msg)
		}
	}
}
