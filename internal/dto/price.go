package dto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// PriceInput accepts either JSON string ("890.00") or JSON number (890.00).
type PriceInput string

func (p PriceInput) String() string {
	return string(p)
}

func (p *PriceInput) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*p = ""
		return nil
	}

	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		*p = PriceInput(strings.TrimSpace(asString))
		return nil
	}

	var asNumber json.Number
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&asNumber); err == nil {
		*p = PriceInput(strings.TrimSpace(asNumber.String()))
		return nil
	}

	return fmt.Errorf("price must be string or number")
}
