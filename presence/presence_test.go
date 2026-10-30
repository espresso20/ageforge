package presence

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

func TestFrameRoundTrip(t *testing.T) {
	in := []byte(`{"a":1}`)
	buf := encodeFrame(opFrame, in)
	if got := binary.LittleEndian.Uint32(buf[0:4]); got != 1 {
		t.Errorf("opcode bytes read %d", got)
	}
	if got := binary.LittleEndian.Uint32(buf[4:8]); got != uint32(len(in)) {
		t.Errorf("length bytes read %d", got)
	}
	f, err := readFrame(bytes.NewReader(buf))
	if err != nil || f.Op != opFrame || !bytes.Equal(f.Data, in) {
		t.Fatalf("round trip: %+v %v", f, err)
	}
	// A length past the cap is refused, not allocated.
	big := make([]byte, 8)
	binary.LittleEndian.PutUint32(big[4:8], maxPayload+1)
	if _, err := readFrame(bytes.NewReader(big)); err == nil {
		t.Error("an oversize frame was read")
	}
	// A short frame is an error.
	if _, err := readFrame(bytes.NewReader(buf[:10])); err == nil {
		t.Error("a short frame was read")
	}
}

func TestHandshakeBytes(t *testing.T) {
	got := handshakeFrame("123")
	want := encodeFrame(0, []byte(`{"client_id":"123","v":1}`))
	if !bytes.Equal(got, want) {
		t.Errorf("handshake = %q, want %q", got, want)
	}
}

func TestActivityFrame(t *testing.T) {
	a := &Activity{Details: "d", State: "s", LargeImage: "icon", LargeText: "AgeForge", Start: time.Unix(1700000000, 0)}
	f, err := readFrame(bytes.NewReader(activityFrame(42, a, "n1")))
	if err != nil || f.Op != opFrame {
		t.Fatalf("frame: %+v %v", f, err)
	}
	var m struct {
		Cmd   string `json:"cmd"`
		Nonce string `json:"nonce"`
		Args  struct {
			PID      int `json:"pid"`
			Activity struct {
				Details    string `json:"details"`
				State      string `json:"state"`
				Assets     map[string]string
				Timestamps map[string]int64
			} `json:"activity"`
		} `json:"args"`
	}
	if err := json.Unmarshal(f.Data, &m); err != nil {
		t.Fatal(err)
	}
	if m.Cmd != "SET_ACTIVITY" || m.Nonce != "n1" || m.Args.PID != 42 || m.Args.Activity.Details != "d" || m.Args.Activity.State != "s" ||
		m.Args.Activity.Assets["large_image"] != "icon" || m.Args.Activity.Timestamps["start"] != 1700000000 {
		t.Errorf("command = %s", f.Data)
	}
	// Cleared: no activity key at all.
	f, _ = readFrame(bytes.NewReader(activityFrame(42, nil, "n2")))
	if strings.Contains(string(f.Data), "activity") {
		t.Errorf("a clear carries an activity: %s", f.Data)
	}
	// No state: the key is left out.
	f, _ = readFrame(bytes.NewReader(activityFrame(42, &Activity{Details: "d"}, "n3")))
	if strings.Contains(string(f.Data), `"state"`) {
		t.Errorf("an empty state was sent: %s", f.Data)
	}
}

func TestFromState(t *testing.T) {
	st := game.NewGameEngine().GetState()
	st.Seed = 77
	st.EpochName = "Fallback Epoch"
	started := time.Unix(1700000000, 0)
	a := FromState(st, started)
	view := st.AdvanceView()
	if want := view.AgeName + " · " + view.YearText; a.Details != want {
		t.Errorf("details = %q, want %q", a.Details, want)
	}
	town := mapmodel.TownNameOf(77)
	if !strings.HasPrefix(a.State, town+" · ") || strings.Contains(a.State, "Prestige") {
		t.Errorf("state = %q", a.State)
	}
	if a.LargeImage != "icon" || a.LargeText != "AgeForge" || !a.Start.Equal(started) {
		t.Errorf("assets or start wrong: %+v", a)
	}
	// Prestige is added when there is any.
	st.Prestige.Level = 3
	if a = FromState(st, started); !strings.HasSuffix(a.State, "Prestige 3") {
		t.Errorf("state with prestige = %q", a.State)
	}
	// The cap: a name too long drops the prestige tail, and no field passes 128 characters.
	st.EpochName = strings.Repeat("x", 140)
	st.EpochKey = ""
	a = FromState(st, started)
	if strings.Contains(a.State, "Prestige") {
		t.Errorf("prestige kept past the cap: %q", a.State)
	}
	if n := len([]rune(a.State)); n > fieldMax {
		t.Errorf("state is %d characters", n)
	}
}

func TestMenuActivity(t *testing.T) {
	a := MenuActivity(time.Unix(5, 0))
	if a.Details != "In the main menu" || a.State != "" {
		t.Errorf("menu activity = %+v", a)
	}
}

func TestClip(t *testing.T) {
	if got := clip(strings.Repeat("é", 200)); len([]rune(got)) != fieldMax {
		t.Errorf("clipped to %d runes", len([]rune(got)))
	}
	if clip("short") != "short" {
		t.Error("a short text changed")
	}
}
