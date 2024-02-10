package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethpandaops/assertoor/pkg/coordinator/clients"
	"github.com/ethpandaops/assertoor/pkg/coordinator/human-duration"
	"github.com/ethpandaops/assertoor/pkg/coordinator/names"
	"github.com/ethpandaops/assertoor/pkg/coordinator/scheduler"
	checkclientsarehealthy "github.com/ethpandaops/assertoor/pkg/coordinator/tasks/check_clients_are_healthy"
	"github.com/ethpandaops/assertoor/pkg/coordinator/types"
	"github.com/ethpandaops/assertoor/pkg/coordinator/vars"
	"github.com/ethpandaops/assertoor/pkg/coordinator/wallet"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	// create client pool
	clientPool, err := clients.NewClientPool(logger.WithField("pkg", "clients"))
	if err != nil {
		panic(fmt.Errorf("failed initializing client pool: %w", err))
	}

	// add clients to pool
	err = clientPool.AddClient(&clients.ClientConfig{
		Name:         "node1",
		ConsensusURL: "https://beacon.dencun-devnet-12.ethpandaops.io/",
		ExecutionURL: "https://rpc.dencun-devnet-12.ethpandaops.io/",
	})
	if err != nil {
		panic(fmt.Errorf("failed adding client to client pool: %w", err))
	}

	// create services
	services := scheduler.NewServicesProvider(
		clientPool,
		wallet.NewManager(
			clientPool.GetExecutionPool(),
			logger.WithField("pkg", "wallet"),
		),
		names.NewValidatorNames(
			&names.Config{
				Inventory: map[string]string{
					"0-100": "name1",
				},
			},
			logger.WithField("pkg", "names"),
		),
	)

	// create variable bag
	variables := vars.NewVariables(nil)

	// create task scheduler
	taskScheduler := scheduler.NewTaskScheduler(logger.WithField("pkg", "scheduler"), services, variables)

	// add some test tasks
	// task 1: check_clients_are_healthy
	_, err = taskScheduler.AddRootTask(&types.TaskOptions{
		Name: checkclientsarehealthy.TaskName,
		Config: scheduler.GetRawConfig(&checkclientsarehealthy.Config{
			PollInterval: human.Duration{Duration: 5 * time.Second},
		}),
		Timeout: human.Duration{Duration: 5 * time.Minute},
	})
	if err != nil {
		panic(fmt.Errorf("failed adding task 1 to scheduler: %w", err))
	}

	// execute tasks
	err = taskScheduler.RunTasks(context.Background(), 0)
	if err != nil {
		panic(fmt.Errorf("task execution returned error: %w", err))
	}

	logger.Infof("All tasks completed!")
}
