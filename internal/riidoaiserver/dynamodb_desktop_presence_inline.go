package riidoaiserver

import "errors"

func decodeDesktopPresenceInline(current map[string]map[string]string) (AIAgentClientSnapshot, error) {
	raw := []byte(dynamoDBStringValue(current, "snapshot_json"))
	if encoded := dynamoDBStringValue(current, "snapshot_gzip"); encoded != "" {
		var err error
		raw, err = desktopPresenceGunzip(encoded)
		if err != nil {
			return AIAgentClientSnapshot{}, err
		}
	}
	if len(raw) > desktopPresencePartMaxBytes {
		return AIAgentClientSnapshot{}, errors.New("desktop presence legacy snapshot exceeds read limit")
	}
	snapshot, _, err := decodeLegacyDynamoDBAIAgentClientSnapshot(map[string]map[string]string{"snapshot_json": {"S": string(raw)}})
	return snapshot, err
}
