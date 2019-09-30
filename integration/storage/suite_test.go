package storage_test

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cf-operator/integration/environment"
	cmdHelper "code.cloudfoundry.org/cf-operator/testing"
	"k8s.io/client-go/rest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Storage Suite")
}

var (
	env              *environment.Environment
	namespacesToNuke []string
	kubeConfig       *rest.Config
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	kubeConfig, err = environment.KubeConfig()
	if err != nil {
		fmt.Printf("WARNING: failed to get kube config: %v\n", err)
	}

	// Ginkgo node 1 gets to setup the CRDs
	err = environment.ApplyCRDs(kubeConfig)
	if err != nil {
		fmt.Printf("WARNING: failed to apply CRDs: %v", err)
	}

	return []byte{}
}, func([]byte) {
	var err error
	kubeConfig, err = environment.KubeConfig()
	if err != nil {
		fmt.Printf("WARNING: failed to get kube config: %v", err)
	}
})

var _ = BeforeEach(func() {
	env = environment.NewEnvironment(kubeConfig)
	err := env.SetupNamespace()
	if err != nil {
		fmt.Printf("WARNING: failed to setup namespace %s: %v\n", env.Namespace, err)
	}
	namespacesToNuke = append(namespacesToNuke, env.Namespace)

	err = env.StartOperator()
	if err != nil {
		fmt.Printf("WARNING: failed to start operator: %v\n", err)
	}

})

var _ = AfterEach(func() {
	env.Teardown(CurrentGinkgoTestDescription().Failed)
})

var _ = AfterSuite(func() {
	// Nuking all namespaces at the end of the run
	for _, namespace := range namespacesToNuke {
		err := cmdHelper.DeleteNamespace(namespace)
		if err != nil {
			fmt.Printf("WARNING: failed to delete namespace %s: %v\n", namespace, err)
		}
		err = cmdHelper.DeleteWebhooks(namespace)
		if err != nil {
			fmt.Printf("WARNING: failed to delete mutatingwebhookconfiguration in %s: %v\n", namespace, err)
		}
	}
})
