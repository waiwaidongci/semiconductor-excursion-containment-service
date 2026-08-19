# Bug reproduction

Metrology snapshot, count, get, and replacement paths access the shared point map without the registry read lock. Snapshots also expose the internal points slice.

Run the targeted race-enabled metrology test. It reports data races and may terminate with `fatal error: concurrent map iteration and map write`.
