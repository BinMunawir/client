package main

import (
	"context"
	"log"
	"math/rand/v2"
	"strconv"

	"github.com/BinMunawir/maal_business/config"
	"github.com/BinMunawir/maal_business/internal/adapters/temporal"
	businessworkflows "github.com/BinMunawir/maal_business/internal/business/workflows"
	"go.temporal.io/sdk/client"
)

// starter builds the boundary DTO and starts a workflow run. The workflow ID is the run's
// idempotency key (standard §8), so re-running with the same CorrID is de-duplicated by
// Temporal. This example onboards a Business; the access workflows are started the same way.
func main() {
	config.Load()

	c, err := temporal.Dial()
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()

	in := businessworkflows.OnboardInput{
		CorrID:      "onboard-" + strconv.Itoa(rand.IntN(99999999)),
		LegalName:   "Acme Trading Company LLC",
		TradeName:   "Acme",
		CRNumber:    "1010101010",
		LegalForm:   "Company",
		SizeSegment: "sme",
		ServiceTier: "gold",
	}
	opts := client.StartWorkflowOptions{
		ID:        in.CorrID,
		TaskQueue: businessworkflows.TaskQueueBusiness,
	}

	we, err := c.ExecuteWorkflow(context.Background(), opts, businessworkflows.Onboard, in)
	if err != nil {
		log.Fatalf("start Onboard workflow: %v", err)
	}
	log.Printf("started: WorkflowID=%s RunID=%s", we.GetID(), we.GetRunID())

	var out businessworkflows.OnboardOutput
	if err := we.Get(context.Background(), &out); err != nil {
		log.Fatalf("workflow failed: %v", err)
	}
	log.Printf("onboarded business: %+v", out.Business)
}
