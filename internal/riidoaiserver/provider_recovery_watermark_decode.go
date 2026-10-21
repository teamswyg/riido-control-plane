package riidoaiserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

func recoveryWatermarkFromJSON(value string, lastEventSeq int64) (*recoveryWatermark, error) {
	if value == "" {
		return nil, nil
	}
	if len(value) > 8192 {
		return nil, errors.New("provider recovery watermark exceeds budget")
	}
	var watermark recoveryWatermark
	if err := json.Unmarshal([]byte(value), &watermark); err != nil {
		return nil, errors.New("provider recovery watermark is invalid")
	}
	if len(watermark.Receipts) == 0 || len(watermark.Receipts) > maxRecoveryReceipts || watermark.Context != watermark.Context.Sanitized() {
		return nil, errors.New("provider recovery watermark is invalid")
	}
	var previous int64
	for _, receipt := range watermark.Receipts {
		key, keyErr := hex.DecodeString(receipt.KeyHash)
		context, contextErr := hex.DecodeString(receipt.ContextHash)
		if keyErr != nil || contextErr != nil || len(key) != 32 || len(context) != 32 || receipt.Seq <= previous || receipt.Seq > lastEventSeq || receipt.At.IsZero() {
			return nil, errors.New("provider recovery receipt is invalid")
		}
		previous = receipt.Seq
	}
	return &watermark, nil
}
