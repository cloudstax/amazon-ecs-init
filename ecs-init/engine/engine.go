// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package engine

import (
	"errors"
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/cloudstax/amazon-ecs-init/ecs-init/docker"
	"github.com/cloudstax/amazon-ecs-init/ecs-init/exec"
	"github.com/cloudstax/amazon-ecs-init/ecs-init/exec/iptables"
	"github.com/cloudstax/amazon-ecs-init/ecs-init/exec/sysctl"
)

const (
	terminalSuccessAgentExitCode = 0
	terminalFailureAgentExitCode = 5
	upgradeAgentExitCode         = 42
)

// Engine contains methods invoked when ecs-init is run
type Engine struct {
	docker                *docker.Client
	loopbackRouting       loopbackRouting
	credentialsProxyRoute credentialsProxyRoute
}

// New creates an instance of Engine
func New() (*Engine, error) {
	dockerclient, err := docker.NewClient()
	if err != nil {
		return nil, err
	}

	cmdExec := exec.NewExec()
	loopbackRouting, err := sysctl.NewIpv4RouteLocalNet(cmdExec)
	if err != nil {
		return nil, err
	}
	credentialsProxyRoute, err := iptables.NewNetfilterRoute(cmdExec)
	if err != nil {
		return nil, err
	}
	return &Engine{
		docker:                dockerclient,
		loopbackRouting:       loopbackRouting,
		credentialsProxyRoute: credentialsProxyRoute,
	}, nil
}

// PreStart prepares the ECS Agent for starting. It also configures the instance
// to handle credentials requests from containers by rerouting these requests to
// to the ECS Agent's credentials endpoint
func (e *Engine) PreStart() error {
	// Enable use of loopback addresses for local routing purposes
	err := e.loopbackRouting.Enable()
	if err != nil {
		return engineError("could not enable loopback routing", err)
	}
	// Add the rerouting netfilter rule for credentials endpoint
	err = e.credentialsProxyRoute.Create()
	if err != nil {
		return engineError("could not create route to the credentials proxy", err)
	}

	return e.docker.CheckAndLoadImage()
}

// StartSupervised starts the ECS Agent and ensures it stays running, except for terminal errors (indicated by an agent exit code of 5)
func (e *Engine) StartSupervised() error {
	agentExitCode := -1
	for agentExitCode != terminalSuccessAgentExitCode && agentExitCode != terminalFailureAgentExitCode {
		err := e.docker.RemoveExistingAgentContainer()
		if err != nil {
			return engineError("could not remove existing Agent container", err)
		}

		// the old container removed or not exists
		log.Info("Starting Amazon EC2 Container Service Agent")
		agentExitCode, err = e.docker.StartAgent()
		if err != nil {
			return engineError("could not start Agent", err)
		}
		log.Infof("Agent exited with code %d", agentExitCode)
		if agentExitCode == upgradeAgentExitCode {
			err = e.docker.DownloadAgentImage()
			if err != nil {
				log.Error("could not upgrade agent", err)
			}
		}
	}
	if agentExitCode == terminalFailureAgentExitCode {
		return errors.New("agent exited with terminal exit code")
	}
	return nil
}

// PreStop sends commands to Docker to stop the ECS Agent
func (e *Engine) PreStop() error {
	log.Info("Stopping Amazon EC2 Container Service Agent")
	err := e.docker.StopAgent()
	if err != nil {
		return engineError("could not stop Amazon EC2 Container Service Agent", err)
	}
	return nil
}

// PostStop cleans up the credentials endpoint setup by disabling loopback
// routing and removing the rerouting rule from the netfilter table
func (e *Engine) PostStop() error {
	log.Info("Cleaning up the credentials endpoint setup for Amazon EC2 Container Service Agent")
	err := e.loopbackRouting.RestoreDefault()

	// Ignore error from Remove() as the netfilter might never have been
	// addred in the first place
	e.credentialsProxyRoute.Remove()
	return err
}

type _engineError struct {
	err     error
	message string
}

func (e _engineError) Error() string {
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func engineError(message string, err error) _engineError {
	return _engineError{
		message: message,
		err:     err,
	}
}
