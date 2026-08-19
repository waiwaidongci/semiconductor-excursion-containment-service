# Bug reproduction

The containment repository returns live mutable lot state, so a caller can alter persisted tool observations and history. The transition table also permits a detected lot to skip containment and release directly.

Run the targeted containment test to observe the leaked tool map/history and invalid transition behavior. The failure reports a repository snapshot containing the caller's added `IMPLANT-02` tool.
