# Bug reproduction

Quarantine lookup and close errors are formatted without preserving their sentinel identity. Missing holds and ownership conflicts therefore become unclassifiable text and can enter retry logic.

Run the targeted quarantine test. It fails while checking that a missing hold remains `ErrHoldMissing` and is not retryable.
