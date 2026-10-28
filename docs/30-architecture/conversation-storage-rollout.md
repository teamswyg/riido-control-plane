# Conversation storage compatibility baseline

Deploy this reader/writer and retention baseline before comment-root/native-history producers. Both legacy manifests and format-1 immutable chunks are readable. Old writers are fenced by the numeric format marker. A stale replica must reload before a retry. After format-1 records are committed, roll back only to this compatible baseline or a later compatible version; never downgrade the shared snapshot. No SQL migration or table replacement is required.

This baseline does not enable RIID-5891/5892 producers, import, or Agent handoff. Unit tests cover legacy compatibility, chunk integrity, interrupted staging, atomic publication, stale-writer rejection, and retention.
