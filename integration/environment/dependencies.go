package environment

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
)

// StartDependencies starts dependencies of cf-operator, like quarks-job
func (e *Environment) StartDependencies() error {
	// TODO maybe install the helm chart instead? no that would require tiller
	cliPath, err := gexec.Build("code.cloudfoundry.org/quarks-job/cmd")
	// TODO quarks-job docker image required and needs to be passed here:
	cmd := exec.Command(cliPath,
		"-n", e.Namespace,
		"-o", "edwardxiao",
		"-r", "quarks-job",
		"-t", "v0.0.0-0.g06315b6",
	)
	_, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

	return err
}
