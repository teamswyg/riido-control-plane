package riidoaiserver

import (
	"slices"
)

func copyMessageReactions(in []AIAgentMessageReaction) []AIAgentMessageReaction {
	out := make([]AIAgentMessageReaction, len(in))
	for i, reaction := range in {
		out[i] = reaction
		out[i].Actors = slices.Clone(reaction.Actors)
	}
	return out
}

func mutateMessageReaction(in []AIAgentMessageReaction, actor AIAgentMessageReactionActor, emoji string, remove bool) []AIAgentMessageReaction {
	out := copyMessageReactions(in)
	for i := range out {
		if out[i].Emoji != emoji {
			continue
		}
		found := slices.Contains(out[i].Actors, actor)
		if remove {
			out[i].Actors = slices.DeleteFunc(out[i].Actors, func(a AIAgentMessageReactionActor) bool { return a == actor })
		} else if !found {
			out[i].Actors = append(out[i].Actors, actor)
		}
		out[i].Count = len(out[i].Actors)
		if out[i].Count == 0 {
			out = slices.Delete(out, i, i+1)
		}
		return out
	}
	if !remove {
		out = append(out, AIAgentMessageReaction{Emoji: emoji, Count: 1, Actors: []AIAgentMessageReactionActor{actor}})
	}
	return out
}
