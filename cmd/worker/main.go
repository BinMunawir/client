package main

import (
	"log"

	"github.com/BinMunawir/maal_business/config"
	access_activities "github.com/BinMunawir/maal_business/internal/access/activities"
	accessworkflows "github.com/BinMunawir/maal_business/internal/access/workflows"
	"github.com/BinMunawir/maal_business/internal/adapters/temporal"
	business_activities "github.com/BinMunawir/maal_business/internal/business/activities"
	businessworkflows "github.com/BinMunawir/maal_business/internal/business/workflows"
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
	wBiz := worker.New(c, businessworkflows.TaskQueueBusiness, worker.Options{})
	wBiz.RegisterWorkflow(businessworkflows.Onboard)
	wBiz.RegisterActivity(business_activities.Register)
	wBiz.RegisterActivity(business_activities.ProvisionOrg)
	wBiz.RegisterActivity(business_activities.Activate)

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
	log.Printf("worker started on task queues %q and %q", businessworkflows.TaskQueueBusiness, accessworkflows.TaskQueueAccess)

	if err := wAcc.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker run: %v", err)
	}
}
