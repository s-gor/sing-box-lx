//go:build with_sgnet

package sgnet

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"
)

func TestFrameGoldenVectors(t *testing.T) {
	tests := []struct {
		name string
		frame Frame
		want []byte
	}{
		{"open", Frame{Type: TypeOpen, StreamID: 1, Payload: []byte{0xaa}}, []byte{1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0xaa}},
		{"data", Frame{Type: TypeData, Flags: FlagFIN, StreamID: 0x01020304, Payload: []byte("hi")}, []byte{1, 2, 1, 1, 2, 3, 4, 0, 0, 0, 2, 'h', 'i'}},
		{"datagram", Frame{Type: TypeDatagram, StreamID: 7, Payload: []byte{1, 2, 3}}, []byte{1, 3, 0, 0, 0, 0, 7, 0, 0, 0, 3, 1, 2, 3}},
		{"open-result", Frame{Type: TypeOpenResult, StreamID: 1, Payload: []byte{OpenOK}}, []byte{1, 7, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			if err := WriteFrame(&b, tt.frame); err != nil { t.Fatal(err) }
			if !bytes.Equal(b.Bytes(), tt.want) { t.Fatalf("wire=%x want=%x", b.Bytes(), tt.want) }
			got, err := ReadFrame(bytes.NewReader(tt.want))
			if err != nil { t.Fatal(err) }
			if got.Type != tt.frame.Type || got.Flags != tt.frame.Flags || got.StreamID != tt.frame.StreamID || !bytes.Equal(got.Payload, tt.frame.Payload) { t.Fatalf("got=%+v want=%+v", got, tt.frame) }
		})
	}
}

func TestFrameRejectsInvalidWire(t *testing.T) {
	if _, err := ReadFrame(bytes.NewReader([]byte{2, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0})); !errors.Is(err, ErrUnsupportedVersion) { t.Fatalf("version: %v", err) }
	if _, err := ReadFrame(bytes.NewReader([]byte{1, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0})); !errors.Is(err, ErrInvalidType) { t.Fatalf("type: %v", err) }
}

func TestProofGolden(t *testing.T) {
	p := Proof([]byte("key"), []byte("data"))
	if hex.EncodeToString(p[:]) != "5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0" { t.Fatalf("%x", p) }
	if !VerifyProof([]byte("key"), []byte("data"), p[:]) { t.Fatal("verify") }
	if VerifyProof([]byte("bad"), []byte("data"), p[:]) { t.Fatal("accepted bad secret") }
}
