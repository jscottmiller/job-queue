## In-memory job queue

A simple REST service exposing an in-memory job queue.

To run the server, execute `./scripts/run.sh`

To run tests, execute `./scripts/test.sh`

### Implementation notes

The in-memory queue is implemented as a single slice, grown according to the
`append` built-in. The `Queue` type maintains this slice, appending new job
entries on enqueue and dequeuing jobs by manipulating an index into the slice
pointing to the head of the queue. The slice is permitted to grow without
bound, though the it will be compacted if more than half of the entries in the
slice are empty. Compaction should only reduce the length of the slice and not
its capacity, so I'd expect the overall memory usage to stabilize after the
service has run for a while. In the worst case, enqueue operations may cause a
slice allocation until the size stabilizes, and dequeue operations may trigger
a compaction, though on average I would expect both of those operations to
execute in constant time.

To help ensure at-least-once processing, a stale job that was dequeued but not
concluded after more than 5 minutes will return to the queue (the 5 minute
expiration is not a hard limit and will vary depending on an expiration timer).
Assuming that the jobs are idempotent, this provides a simple recovery
mechanism if a processor crashes during execution.

No authorization is implemented - any client can perform operations on any job
they wish. The only authorization that may be worth performing is to ensure
that the client that requests to conclude a job is the same client that
dequeued it.

### Future work

The tests that are implemented are useful but by no means exhaustive. In
particular, more testing of the both the queue compaction and processing
timeout.

Parameters could be added to adjust when to compact the queue slice, as well as
the slice's initial size and growth strategy. Before implementing any of that,
I'd like to add a suite of benchmark tests to 1) ensure that the compaction is
working as I expect with larger workloads and 2) to get a baseline for
performance and see if those parameters are even needed.

Along the same lines, concurrent access of the queue is handled via a single
mutex around all operations. Benchmarking would help uncover the performance
impact of this, and in particular I'd be curious to see how the more expensive
operations (compaction, stale job search) impact concurrent requests.

Finally, the HTTP harness was written quickly and without a lot of bells and
whistles. At minimum, I would want to add better logging and error reporting to
a production service, as well as basic authentication and authorization
middleware.
