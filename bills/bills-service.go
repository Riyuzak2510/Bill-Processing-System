package bills

import (
	"context"
	"fmt"
	"time"

	"encore.app/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Service handles bill operations and Temporal workflow management
//
//encore:service
type Service struct {
	workflows      map[string]string // In-memory storage for demo
	temporalClient client.Client
	workers        []worker.Worker
}

var (
	billsTaskQueue = "local-bills"
	localHost      = "127.0.0.1:7233"
)

// Service initialization
var service *Service

// Initialize the service when the package loads
func init() {
	var err error
	service, err = initService()
	if err != nil {
		// In production, you might want to handle this differently
		panic(fmt.Sprintf("Failed to initialize bills service: %v", err))
	}
}

// GetService returns the initialized service instance
// This allows other packages to access the service if needed
func GetService() *Service {
	return service
}

// Initialize the service with Temporal client
func initService() (*Service, error) {
	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort: localHost, // Temporalite default port
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}
	workers := []worker.Worker{}
	for i := 0; i < 10; i++ {
		worker := worker.New(temporalClient, billsTaskQueue, worker.Options{
			WorkerStopTimeout: 30 * time.Second,
		})
		err = worker.Start()
		if err != nil {
			return nil, fmt.Errorf("failed to start worker: %w", err)
		}
		worker.RegisterWorkflow(workflows.BillWorkflow)
		workers = append(workers, worker)
	}

	return &Service{
		workflows:      make(map[string]string),
		temporalClient: temporalClient,
		workers:        workers,
	}, nil
}

// Shutdown gracefully closes the service
func (s *Service) Shutdown(force context.Context) {
	for _, w := range s.workers {
		w.Stop()
	}
	if s.temporalClient != nil {
		s.temporalClient.Close()
	}
}

// GetTemporalClient returns the Temporal client
func (s *Service) GetTemporalClient() client.Client {
	return s.temporalClient
}

// GetTaskQueue returns the task queue name
func (s *Service) GetTaskQueue() string {
	return billsTaskQueue
}

// GetWorkflowIDForCustomer returns the workflow ID for a given customer
func (s *Service) GetWorkflowIDForCustomer(customerID string) (string, bool) {
	id, found := s.workflows[customerID]
	return id, found
}
