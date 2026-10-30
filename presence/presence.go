// Package presence shows the player's run on their Discord profile while the Discord
// desktop app is running. It talks to the app on the same machine over its local IPC
// socket (a unix socket, or a named pipe on Windows); nothing goes to any server of ours,
// the player never logs in, and when Discord is not running nothing happens and nothing
// is said.
package presence

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// AppID is the Discord application id; empty turns presence off.
const AppID = "1565572290375057540"

// The opcodes of Discord's IPC framing.
const (
	opHandshake uint32 = 0
	opFrame     uint32 = 1
	opClose     uint32 = 2
	opPing      uint32 = 3
	opPong      uint32 = 4
)

// maxPayload bounds a frame we are willing to read; Discord's own are far smaller.
const maxPayload = 1 << 20

// Activity is what the profile shows. A nil *Activity means show nothing.
type Activity struct {
	Details    string
	State      string
	LargeImage string
	LargeText  string
	// Start is when the game began; Discord shows the time since. Zero shows none.
	Start time.Time
}

// wire is the activity as Discord's SET_ACTIVITY takes it.
func (a *Activity) wire() map[string]any {
	out := map[string]any{"details": a.Details}
	if a.State != "" {
		out["state"] = a.State
	}
	if a.LargeImage != "" || a.LargeText != "" {
		assets := map[string]string{}
		if a.LargeImage != "" {
			assets["large_image"] = a.LargeImage
		}
		if a.LargeText != "" {
			assets["large_text"] = a.LargeText
		}
		out["assets"] = assets
	}
	if !a.Start.IsZero() {
		out["timestamps"] = map[string]int64{"start": a.Start.Unix()}
	}
	return out
}

// frame is one message of the protocol.
type frame struct {
	Op   uint32
	Data []byte
}

// encodeFrame lays a frame out as Discord reads it: a little-endian uint32 opcode, a
// little-endian uint32 length, then the JSON.
func encodeFrame(op uint32, payload []byte) []byte {
	var head [8]byte
	binary.LittleEndian.PutUint32(head[0:4], op)
	binary.LittleEndian.PutUint32(head[4:8], uint32(len(payload)))
	return append(head[:], payload...)
}

// readFrame reads one frame.
func readFrame(r io.Reader) (frame, error) {
	var head [8]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return frame{}, err
	}
	n := binary.LittleEndian.Uint32(head[4:8])
	if n > maxPayload {
		return frame{}, fmt.Errorf("presence: frame of %d bytes is too large", n)
	}
	data := make([]byte, n)
	if _, err := io.ReadFull(r, data); err != nil {
		return frame{}, err
	}
	return frame{Op: binary.LittleEndian.Uint32(head[0:4]), Data: data}, nil
}

// handshakeFrame is the first frame sent: protocol version 1 and the application id.
func handshakeFrame(appID string) []byte {
	payload, _ := json.Marshal(map[string]any{"v": 1, "client_id": appID})
	return encodeFrame(opHandshake, payload)
}

// activityFrame is a SET_ACTIVITY command for process pid. A nil activity clears it.
func activityFrame(pid int, a *Activity, nonce string) []byte {
	args := map[string]any{"pid": pid}
	if a != nil {
		args["activity"] = a.wire()
	}
	payload, _ := json.Marshal(map[string]any{"cmd": "SET_ACTIVITY", "args": args, "nonce": nonce})
	return encodeFrame(opFrame, payload)
}
