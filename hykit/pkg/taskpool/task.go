package taskpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"

	"code.jshyjdtech.com/godev/hykit/log"
)

type TaskPool struct {
	wg     sync.WaitGroup
	cancel context.CancelFunc

	logger log.Logger
	conf   config.Config

	acceptJobChan chan IJob
	pool          chan chan IJob
	works         []*worker
	maxWorkers    int64
	runWorkers    int64
}

type Option func(*TaskPool)

var (
	once       sync.Once
	TaskClient *TaskPool
)

const defaultPoolSize = 5

func NewTaskPool(opts ...Option) *TaskPool {
	once.Do(func() {
		TaskClient = &TaskPool{
			//pool:       make(chan chan IJob, defaultPoolSize),
			acceptJobChan: make(chan IJob),

			works:      make([]*worker, 0),
			maxWorkers: defaultPoolSize,

			wg: sync.WaitGroup{},
		}

		for _, opt := range opts {
			opt(TaskClient)
		}

		if TaskClient.logger == nil {
			TaskClient.logger = log.NewLogger()
		}

		if TaskClient.conf == nil {
			TaskClient.conf = config.NewMemConfig()
		}

		poolSize := TaskClient.conf.GetInt64("taskpool_max_count")
		if 0 < poolSize && poolSize < 1000 {
			TaskClient.maxWorkers = poolSize
		}

		TaskClient.pool = make(chan chan IJob, TaskClient.maxWorkers)

		TaskClient.logger.Infof("开启了[%d]task", TaskClient.maxWorkers)

	})
	return TaskClient
}

func (t *TaskPool) WithTaskSize(taskSize int64) {
	TaskClient.maxWorkers = taskSize
	TaskClient.pool = make(chan chan IJob, taskSize)
}

func WithTaskLogger(log log.Logger) Option {
	return func(task *TaskPool) {
		task.logger = log
	}
}

func WithTaskConf(conf config.Config) Option {
	return func(task *TaskPool) {
		task.conf = conf
	}
}

func (t *TaskPool) Concurrency() int64 {
	return t.maxWorkers - int64(len(t.pool))
}

func (t *TaskPool) AddJobs(jobs ...IJob) {
	for _, job := range jobs {
		t.acceptJobChan <- job
	}
}

func (t *TaskPool) AddFunc(f func()) {
	var jobFunc JobCall = f
	t.acceptJobChan <- jobFunc
}

func GetTaskClient() *TaskPool {
	if TaskClient == nil {
		TaskClient = NewTaskPool()
	}

	return TaskClient
}

func (t *TaskPool) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	// starting n number of workers
	var i int64
	for i = 0; i < t.maxWorkers; i++ {
		worker := newWorker(i, t.pool)
		t.wg.Add(1)
		go func(idx int64) {
			defer t.wg.Done()
			t.logger.Infof("worker[%v], 开始执行....", idx+1)
			atomic.AddInt64(&t.runWorkers, 1)
			worker.start()
			t.logger.Infof("worker[%v], 结束执行....", idx+1)
		}(i)
		t.works = append(t.works, worker)
	}
	//wait work start
	for {
		if atomic.LoadInt64(&t.runWorkers) == t.maxWorkers {
			t.logger.Infof("work start success;")
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// accept task
	go t.process(ctx)
}

func (t *TaskPool) process(ctx context.Context) {
	for {
		select {
		case job := <-t.acceptJobChan:
			// a job request has been received
			// 直接在当前routine完成到worker的分配*/
			jobChannel, ok := <-t.pool
			if !ok {
				t.logger.Infof("Task failure")
				return
			}

			// dispatch the job to the worker job channel
			jobChannel <- job
		case <-ctx.Done():
			var i int64
			for i = 0; i < t.maxWorkers; i++ {
				jobChannel := <-t.pool
				close(jobChannel)
			}
			close(t.pool)

			t.logger.Infof("congratulations, Task over")
			return
		}
	}
}

func (t *TaskPool) Stop() {
	t.cancel()
	t.wg.Wait()

	for _, w := range t.works {
		w.stop()
	}
}
