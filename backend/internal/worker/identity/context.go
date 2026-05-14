package identity

import "sync/atomic"

var orgID atomic.Uint64
var workerID atomic.Uint64

// Set records process-wide worker identity.
func Set(org, worker uint) {
	orgID.Store(uint64(org))
	workerID.Store(uint64(worker))
}

// OrgID returns the organization ID bound to this worker process.
func OrgID() uint {
	return uint(orgID.Load())
}

// WorkerID returns the worker ID bound to this worker process.
func WorkerID() uint {
	return uint(workerID.Load())
}
