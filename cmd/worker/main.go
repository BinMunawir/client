package main

import (
	"log"

	"github.com/BinMunawir/client/config"
	access_activities "github.com/BinMunawir/client/internal/access/activities"
	accessworkflows "github.com/BinMunawir/client/internal/access/workflows"
	"github.com/BinMunawir/client/internal/adapters/temporal"
	activities "github.com/BinMunawir/client/internal/client/activities"
	workflows "github.com/BinMunawir/client/internal/client/workflows"
	"go.temporal.io/sdk/worker"
)

// The worker runtime executes the durable flows. This service has two capabilities on two
// task queues, so the process hosts one worker per queue: the business worker runs in the
// background and the access worker blocks on the interrupt channel.
func main() {
	config.Load()

	c, err := temporal.Dial()
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()

	// business capability
	wBiz := worker.New(c, workflows.TaskQueueBusiness, worker.Options{})
	wBiz.RegisterWorkflow(workflows.Onboard)
	wBiz.RegisterActivity(activities.Register)
	wBiz.RegisterActivity(activities.ProvisionOrg)
	wBiz.RegisterActivity(activities.Activate)

	// access capability
	wAcc := worker.New(c, accessworkflows.TaskQueueAccess, worker.Options{})
	wAcc.RegisterWorkflow(accessworkflows.Invite)
	wAcc.RegisterWorkflow(accessworkflows.AcceptInvitation)
	wAcc.RegisterWorkflow(accessworkflows.ProvisionServiceAccount)
	wAcc.RegisterActivity(access_activities.IssueInvitation)
	wAcc.RegisterActivity(access_activities.ValidateInvitation)
	wAcc.RegisterActivity(access_activities.ProvisionUser)
	wAcc.RegisterActivity(access_activities.MaterializeMembership)
	wAcc.RegisterActivity(access_activities.AcceptInvite)
	wAcc.RegisterActivity(access_activities.CreateServiceAccount)

	if err := wBiz.Start(); err != nil {
		log.Fatalf("start business worker: %v", err)
	}
	defer wBiz.Stop()
	log.Printf("worker started on task queues %q and %q", workflows.TaskQueueBusiness, accessworkflows.TaskQueueAccess)

	if err := wAcc.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker run: %v", err)
	}
}
