package taskpool

// worker represents the worker that executes the job
type worker struct {
	id      int64
	pool    chan chan IJob
	jobChnl chan IJob
	quit    chan bool
}

func newWorker(id int64, p chan chan IJob) *worker {
	return &worker{
		id:      id,
		pool:    p,
		jobChnl: make(chan IJob),
		quit:    make(chan bool),
	}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *worker) start() {
	for {
		// register the current worker into the worker queue.
		w.pool <- w.jobChnl
		select {
		case job, ok := <-w.jobChnl:
			if !ok { //channel已经关闭
				return
			}
			// we have received a work request.
			job.Run()
		case <-w.quit:
			// we have received a signal to stop
			close(w.jobChnl)
			return
		}
	}
}

// Stop signals the worker to stop listening for work requests.
func (w *worker) stop() {
	go func() {
		w.quit <- true
	}()
}
