# Bug reproduction

Review storage and service methods discard the caller context. Canceled votes and waits continue, deadlines are ignored, and state can be written after the request has ended.

Run the targeted review test with a canceled context. It reports that the store lookup succeeds instead of returning `context.Canceled`.
