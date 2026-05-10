package worker

import (
	"github.com/insmtx/Leros/backend/internal/worker/client"
)

type Worker = client.WorkerClient
type WorkerConfig = client.WorkerConfig

var NewWorker = client.NewWorker
