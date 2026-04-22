package netproto

import (
	"encoding/json"
	"fmt"
	"io"
)

type Codec struct {
	enc *json.Encoder
	dec *json.Decoder
}

func NewCodec(r io.Reader, w io.Writer) *Codec {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &Codec{enc: enc, dec: json.NewDecoder(r)}
}

func (c *Codec) Write(msg Envelope) error {
	if err := c.enc.Encode(msg); err != nil {
		return fmt.Errorf("encode envelope: %w", err)
	}
	return nil
}

func (c *Codec) Read() (Envelope, error) {
	var env Envelope
	if err := c.dec.Decode(&env); err != nil {
		return Envelope{}, fmt.Errorf("decode envelope: %w", err)
	}
	return env, nil
}

func DecodePayload[T any](env Envelope) (T, error) {
	var zero T
	raw, err := json.Marshal(env.Payload)
	if err != nil {
		return zero, fmt.Errorf("marshal payload: %w", err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, fmt.Errorf("unmarshal payload: %w", err)
	}
	return out, nil
}
