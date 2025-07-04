package utils

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Task represents a unit of work to be processed by the worker pool
type Task interface {
	Execute(ctx context.Context) error
	ID() string
	Name() string
}

// WorkerPool manages a pool of workers for concurrent task processing
type WorkerPool struct {
	numWorkers int
	taskQueue  chan Task
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.Logger
	metrics    *MetricsCollector
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int, queueSize int, logger *zap.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		numWorkers: numWorkers,
		taskQueue:  make(chan Task, queueSize),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		metrics:    GetGlobalMetricsCollector(logger),
	}

	// Initialize metrics
	pool.metrics.Counter("worker_pool_tasks_submitted", nil)
	pool.metrics.Counter("worker_pool_tasks_completed", nil)
	pool.metrics.Counter("worker_pool_tasks_failed", nil)
	pool.metrics.Gauge("worker_pool_queue_size", nil)
	pool.metrics.Gauge("worker_pool_active_workers", nil)

	return pool
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.logger.Info("Starting worker pool", zap.Int("workers", wp.numWorkers))
	
	// Set initial active workers gauge
	wp.metrics.Gauge("worker_pool_active_workers", nil).Set(0)
	
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop stops the worker pool and waits for all workers to finish
func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool")
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	wp.logger.Info("Worker pool stopped")
}

// Submit adds a task to the queue
func (wp *WorkerPool) Submit(task Task) bool {
	select {
	case <-wp.ctx.Done():
		return false
	case wp.taskQueue <- task:
		wp.metrics.Counter("worker_pool_tasks_submitted", nil).Inc()
		wp.metrics.Gauge("worker_pool_queue_size", nil).Set(float64(len(wp.taskQueue)))
		return true
	}
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	wp.logger.Debug("Worker started", zap.Int("worker_id", id))
	wp.metrics.Gauge("worker_pool_active_workers", nil).Add(1)
	
	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Debug("Worker stopping due to context cancellation", zap.Int("worker_id", id))
			wp.metrics.Gauge("worker_pool_active_workers", nil).Add(-1)
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				wp.logger.Debug("Worker stopping due to closed queue", zap.Int("worker_id", id))
				wp.metrics.Gauge("worker_pool_active_workers", nil).Add(-1)
				return
			}
			
			wp.metrics.Gauge("worker_pool_queue_size", nil).Set(float64(len(wp.taskQueue)))
			wp.processTask(task, id)
		}
	}
}

// processTask executes a task and handles any errors
func (wp *WorkerPool) processTask(task Task, workerID int) {
	taskLogger := wp.logger.With(
		zap.String("task_id", task.ID()),
		zap.String("task_name", task.Name()),
		zap.Int("worker_id", workerID),
	)
	
	taskLogger.Debug("Processing task")
	
	// Create a timer to track task execution time
	timer := wp.metrics.Histogram("worker_pool_task_duration_ms", map[string]string{
		"task_name": task.Name(),
	}).Timer()
	
	// Execute the task
	err := task.Execute(wp.ctx)
	
	// Stop the timer
	timer()
	
	if err != nil {
		taskLogger.Error("Task execution failed", zap.Error(err))
		wp.metrics.Counter("worker_pool_tasks_failed", map[string]string{
			"task_name": task.Name(),
		}).Inc()
	} else {
		taskLogger.Debug("Task completed successfully")
		wp.metrics.Counter("worker_pool_tasks_completed", map[string]string{
			"task_name": task.Name(),
		}).Inc()
	}
}

// QueueSize returns the current number of tasks in the queue
func (wp *WorkerPool) QueueSize() int {
	return len(wp.taskQueue)
}

// QueueCapacity returns the capacity of the task queue
func (wp *WorkerPool) QueueCapacity() int {
	return cap(wp.taskQueue)
}

// WorkerCount returns the number of workers in the pool
func (wp *WorkerPool) WorkerCount() int {
	return wp.numWorkers
}