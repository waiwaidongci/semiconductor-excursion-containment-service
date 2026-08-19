# Bug reproduction

Sampling replaces the caller's context with a background context. Canceled requests consequently lose their cancellation before serial or parallel equipment reads begin.

Run the targeted sampling test with an already-canceled context. It reports that cancellation is not visible through the request contract.
