package dispatcher

var (
	MaxQueue = 100
	// A buffered channel that we can send work requests on.
	JobQueue chan Job
)

type Job interface {
	Do() error
}

type CommonJob struct {
	Args   []interface{}
	DoFunc func(args ...interface{}) error
	Done   chan error
}

func (job *CommonJob) Do() error {
	err := job.DoFunc(job.Args...)
	if job.Done != nil {
		job.Done <- err
	}
	return err
}

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				if err := job.Do(); err != nil {

				}

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	w.quit <- true
}

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	maxWorkers int
	WorkerPool chan chan Job
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	JobQueue = make(chan Job, MaxQueue)
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			//go func(job Job) {
			// try to obtain a worker job channel that is available.
			// this will block until a worker is idle
			jobChannel := <-d.WorkerPool

			// dispatch the job to the worker job channel
			jobChannel <- job
			//}(job)
		}
	}
}
