package temporal

import (
	"context"
	"fmt"
	"log"

	"github.com/BinMunawir/maal_business/config"
	"go.temporal.io/sdk/client"
)

// Dial connects to the Temporal server and verifies it is reachable. Both the worker
// (cmd/worker) and the workflow starter (cmd/starter) use it (service standard §7). Host
// and namespace come from config, defaulting to the local Temporal dev server.
func Dial() (client.Client, error) {
	hostPort := config.CNF.Temporal.HostPort
	if hostPort == "" {
		hostPort = client.DefaultHostPort
	}
	namespace := config.CNF.Temporal.Namespace
	if namespace == "" {
		namespace = client.DefaultNamespace
	}

	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal: dial: %w", err)
	}
	if _, err := c.CheckHealth(context.Background(), &client.CheckHealthRequest{}); err != nil {
		c.Close()
		return nil, fmt.Errorf("temporal: health check: %w", err)
	}
	log.Printf("temporal connected on %s (namespace %q)", hostPort, namespace)
	return c, nil
}
