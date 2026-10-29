package riidoaiserver

// runReadOwner is the work, meeting or Agent list a child read belongs to.
type runReadOwner struct{ kind, id string }

const runReadOwnerAgentDirectory = "agent_directory"

// setRunReadOwner records which visible work or meeting a child source belongs
// to. Only already-visible keys change, so no hidden source becomes visible.
func setRunReadOwner(metadata map[runReadSourceKey]runReadMetadata, source AssignmentContextSource, owner runReadOwner) {
	key := runReadSourceKey{source.Kind, source.ID, source.Part}
	value, visible := metadata[key]
	if !visible || owner.kind == "" || (owner.kind == source.Kind && owner.id == source.ID) {
		return
	}
	value.owner = owner
	metadata[key] = value
}
