# Bug reproduction

Investigation workers register with the wait group inside their goroutines, while the coordinator closes the results channel immediately. Workers can send after close and the report can lose findings.

Run the targeted race-enabled investigation test. It reproduces `panic: send on closed channel` from the chamber worker result send.
