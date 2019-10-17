package environment

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"code.cloudfoundry.org/cf-operator/pkg/kube/operator"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	helper "code.cloudfoundry.org/quarks-utils/testing/testhelper"
)

// StartOperator prepares and starts the cf-operator
func (e *Environment) StartOperator() error {
	mgr, err := e.setupCFOperator()
	if err != nil {
		return errors.Wrapf(err, "Setting up CF Operator failed.")
	}

	e.Stop = e.startOperator(mgr)
	err = helper.WaitForPort(
		"127.0.0.1",
		strconv.Itoa(int(e.Config.WebhookServerPort)),
		1*time.Minute)
	if err != nil {
		return errors.Wrapf(err, "Integration setup failed. Waiting for port %d failed.", e.Config.WebhookServerPort)
	}

	return nil
}

func (e *Environment) setupCFOperator() (manager.Manager, error) {
	var err error
	whh, found := os.LookupEnv("CF_OPERATOR_WEBHOOK_SERVICE_HOST")
	if !found {
		return nil, errors.Errorf("Please set CF_OPERATOR_WEBHOOK_SERVICE_HOST to the host/ip the operator runs on and try again")
	}
	e.Config.WebhookServerHost = whh

	port, err := getWebhookServicePort(e.ID)
	if err != nil {
		return nil, err
	}

	sshUser, shouldForwardPort := os.LookupEnv("ssh_server_user")
	if shouldForwardPort {
		go func() {
			cmd := exec.Command(
				"ssh", "-fNT", "-i", "/tmp/cf-operator-tunnel-identity", "-o",
				"UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", "-R",
				fmt.Sprintf("%s:%[2]d:localhost:%[2]d", whh, port),
				fmt.Sprintf("%s@%s", sshUser, whh))

			stdOutput, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("SSH TUNNEL FAILED: %s\nOUTPUT: %s", err.Error(), string(stdOutput))
			}
		}()
	}

	e.Config.WebhookServerPort = port

	e.Config.OperatorNamespace = e.Namespace
	e.Config.Namespace = e.Namespace

	dockerImageOrg, found := os.LookupEnv("DOCKER_IMAGE_ORG")
	if !found {
		dockerImageOrg = "cfcontainerization"
	}

	dockerImageRepo, found := os.LookupEnv("DOCKER_IMAGE_REPOSITORY")
	if !found {
		dockerImageRepo = "cf-operator"
	}

	dockerImageTag, found := os.LookupEnv("DOCKER_IMAGE_TAG")
	if !found {
		return nil, errors.Errorf("required environment variable DOCKER_IMAGE_TAG not set")
	}

	err = config.SetupOperatorDockerImage(dockerImageOrg, dockerImageRepo, dockerImageTag)
	if err != nil {
		return nil, err
	}
	fmt.Printf("cf-operator docker image: %s", config.GetOperatorDockerImage())

	ctx := e.SetupLoggerContext("cf-operator-tests")

	mgr, err := operator.NewManager(ctx, e.Config, e.KubeConfig, manager.Options{
		Namespace:          e.Namespace,
		MetricsBindAddress: "0",
		LeaderElection:     false,
		Port:               int(e.Config.WebhookServerPort),
		Host:               "0.0.0.0",
	})

	return mgr, err
}

func (e *Environment) startOperator(mgr manager.Manager) chan struct{} {
	stop := make(chan struct{})
	go func() {
		err := mgr.Start(stop)
		if err != nil {
			panic(err)
		}
	}()
	return stop
}

func getWebhookServicePort(namespaceCounter int) (int32, error) {
	port := int64(40000)
	portString, found := os.LookupEnv("CF_OPERATOR_WEBHOOK_SERVICE_PORT")
	if found {
		var err error
		port, err = strconv.ParseInt(portString, 10, 32)
		if err != nil {
			return -1, errors.Wrapf(err, "Parsing portSting %s failed", portString)
		}
	}
	return int32(port) + int32(namespaceCounter), nil
}
