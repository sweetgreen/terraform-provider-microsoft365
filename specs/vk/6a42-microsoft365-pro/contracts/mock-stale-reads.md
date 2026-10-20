# Contract: group mock stale reads (`mocks/responders.go`)

```go
// SetStaleReadsAfterPatch makes the next n GETs of a group after each PATCH return the group as
// it was before that PATCH. n = -1 means it never catches up. n = 0 (the default) turns it off.
func (m *GroupMock) SetStaleReadsAfterPatch(n int)
```

- PATCH: if the switch is on, snapshot a copy of the group before merging the body, and set
  that group's remaining stale count to n.
- GET of a group with a remaining count other than 0 and a snapshot: serve the snapshot, and
  decrement the count when it's positive.
- `RegisterMocks` and `CleanupMockState` reset the switch, the snapshots and the counts.
