# Refactor & Coverage TODOs

## internal/config
- [x] Add tests that cover config layering (global fallback vs. repo vs. explicit file) including unreadable/missing global configs.
- [x] Document and assert include/merge behaviour so list overwrites don’t regress.

## internal/core
- [x] Add streaming aggregator coverage for the zero-token fallback path.
- [x] Exercise tool executor branches: notifier selection, JSON validation failures, terminal-only tool exits.

## internal/team
- [x] Write integration coverage for delegation sessions (missing agent spawn, `AGENTRY_DELEGATION_TIMEOUT`, workspace context injection, shared-memory updates).
- [x] Add concurrency tests for task orchestration (`AssignTask`, `ExecuteParallelTasks`) and decide on locking vs. channel handoff.
- [x] Clarify expectations for coordination helpers (`PublishWorkspaceEvent`, history trimming, `checkWorkCompleted`) and test the heuristics.

## Dead / Disconnected Code Candidates
- (none outstanding)
