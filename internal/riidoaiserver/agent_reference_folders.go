package riidoaiserver

import (
	"errors"
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func (s *DevelopmentAIAgentClientStore) normalizeReferenceFoldersLocked(folders *assignmentcontract.ReferenceFolders, runtimeID string) (*assignmentcontract.ReferenceFolders, error) {
	normalized, err := folders.Normalize()
	if err != nil || normalized == nil {
		return normalized, err
	}
	device, _, ok := s.deviceRuntimeByRuntimeIDLocked(runtimeID)
	if !ok || !normalized.Matches(device.DeviceID, runtimeID) {
		return nil, errors.New("reference_folders must match the selected device and runtime")
	}
	if len(normalized.Paths) == 0 {
		return nil, nil
	}
	return normalized, nil
}

func (s *Store) normalizeAssignmentReferenceFolders(req AssignRequest) (*assignmentcontract.ReferenceFolders, error) {
	folders, err := req.ReferenceFolders.Normalize()
	if err != nil || folders == nil {
		return folders, err
	}
	if s.agentRegistry == nil {
		return nil, errors.New("reference_folders requires a runtime binding")
	}
	binding, ok := s.agentRegistry.LookupAgent(req.AgentID)
	if !ok || !folders.Matches(binding.DeviceID, binding.RuntimeID) {
		return nil, errors.New("reference_folders runtime binding mismatch")
	}
	return folders, nil
}

func promptWithReferenceFolders(prompt string, folders *assignmentcontract.ReferenceFolders) string {
	if instruction := folders.Instruction(); instruction != "" {
		return prompt + "\n\n" + instruction
	}
	return prompt
}
