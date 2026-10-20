# Contextual task progress — RIID-5860

User-visible progress describes actual work on the current task. The daemon's
existing structured progress codes supply a short `description` (1101/1103) or
`summary` (1104); no additional model call or execution state is introduced.

The server projects only those known arguments into optional `display_message`
on SSE/legacy progress lines and v3 history progress rows. The public response
does not expose raw progress arguments. Unknown/unclassified output, assistant
drafts, lifecycle/queue announcements, empty or oversized descriptions do not
gain a display message. Existing raw message/body fields, sequence, assignment
identity, final results and stop/follow-up behavior stay compatible.

Descriptions are plain text, one line, at most 180 Unicode characters. The daemon
instruction asks for one task-specific sentence in the user's language, with no
internal tool transcript, repeated status or unobserved success claims. This is
a presentation boundary, not an authorization or execution decision.

The client uses `display_message` for recent progress and expandable history;
generic waiting/working/responding labels are fallback only. Older clients and
older daemons remain compatible, but old unclassified logs cannot be converted
into specific progress by guessing their meaning.

Persisted thread records and outbound SSE replay events restore the saved
projection so a server restart does not erase contextual history. Ingest line
decoding ignores incoming `display_message`; normalization clears any restored
projection before accepting a new batch. Only classified arguments can create
a new public summary. History caches include the projected summary in their
identity to avoid reusing stale descriptions.

Verify JSON projection for live and cold history, unknown/malformed inputs,
unchanged private metadata, and distinct descriptions across different tasks.
Server deployment is required; improved report quality also requires the daemon
instruction update. Local fixtures are not proof of a real provider run.
