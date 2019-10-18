package converter

import (
	corev1 "k8s.io/api/core/v1"

	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

// operatorDockerImage is the location of the operators own docker image
var operatorDockerImage = ""

// SetupOperatorDockerImage initializes the package scoped variable
func SetupOperatorDockerImage(org, repo, tag, pullPolicy string) error {
	var err error
	operatorDockerImage, err = names.GetDockerSourceName(org, repo, tag)
	if err != nil {
		return err
	}

	// setup quarks job docker image, too.
	// will have to change this once the persist command is moved to quarks-job
	err = config.SetupOperatorDockerImage(org, repo, tag)
	if err != nil {
		return err
	}
	return config.SetupOperatorImagePullPolicy(pullPolicy)
}

// GetOperatorDockerImage returns the image name of the operator docker image
func GetOperatorDockerImage() string {
	return operatorDockerImage
}

// GetOperatorImagePullPolicy returns the image pull policy to be used for generated pods
func GetOperatorImagePullPolicy() corev1.PullPolicy {
	return config.GetOperatorImagePullPolicy()
}
