# Bug reproduction

Evidence processing defers every resource close until the batch returns, overwrites processing errors with close errors, and commits a failed transaction. Result IDs also retain a caller-visible slice.

Run the targeted evidence test. It reports a maximum of twelve simultaneously open resources instead of one and an incorrect transaction lifecycle.
