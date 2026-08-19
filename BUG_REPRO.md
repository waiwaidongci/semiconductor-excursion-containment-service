# Bug reproduction

Supplier pending and missing errors lose their sentinel identity while crossing the gateway and service layers. Retry classification then disagrees with the request status persisted in the store.

Run the targeted supplier test. It fails while checking that pending evidence remains retryable and that the request is stored as pending.
