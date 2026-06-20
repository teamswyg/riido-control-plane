package riidoaiserver

func (s *DynamoDBAssignmentOperationStore) Close() error {
	if s == nil {
		return nil
	}
	reply := make(chan error, 1)
	select {
	case s.commands <- dynamoDBAssignmentOperationCommand{close: true, reply: reply}:
		return <-reply
	case <-s.done:
		return nil
	}
}
