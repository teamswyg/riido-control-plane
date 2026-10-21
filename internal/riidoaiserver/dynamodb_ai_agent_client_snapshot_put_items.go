package riidoaiserver

import "context"

func (s *DynamoDBAIAgentClientSnapshot) putSnapshotPartItemsSequentially(ctx context.Context, items []map[string]map[string]string, credentials AWSCredentials) error {
	for _, item := range items {
		if err := s.putSnapshotItem(ctx, item, credentials); err != nil {
			return err
		}
	}
	return nil
}
