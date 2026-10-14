package riidoaiserver

import "errors"

func (s *DynamoDBAIAgentClientSnapshot) loop() {
	defer close(s.done)
	var cachedCredentials AWSCredentials
	for cmd := range s.commands {
		if cmd.close {
			cmd.errDone <- nil
			return
		}
		s.handleCommand(cmd, &cachedCredentials)
	}
}

func (s *DynamoDBAIAgentClientSnapshot) handleCommand(cmd dynamoDBAIAgentClientSnapshotCommand, cachedCredentials *AWSCredentials) {
	credentials, err := cachedAWSCredentials(cmd.ctx, s.now, s.credentialsProvider, cachedCredentials)
	if err != nil {
		s.replyCommandError(cmd, err)
		return
	}
	if cmd.load {
		snapshot, ok, err := s.load(cmd.ctx, credentials)
		cmd.loadDone <- dynamoDBAIAgentClientSnapshotLoadResult{snapshot: snapshot, ok: ok, err: err}
		return
	}
	if cmd.save == nil {
		cmd.errDone <- errors.New("riidoaiserver: nil DynamoDB AI Agent client snapshot")
		return
	}
	cmd.errDone <- s.save(cmd.ctx, *cmd.save, credentials)
}

func (s *DynamoDBAIAgentClientSnapshot) replyCommandError(cmd dynamoDBAIAgentClientSnapshotCommand, err error) {
	if cmd.load {
		cmd.loadDone <- dynamoDBAIAgentClientSnapshotLoadResult{err: err}
		return
	}
	cmd.errDone <- err
}
