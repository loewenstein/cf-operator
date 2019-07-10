package manifest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bdc "code.cloudfoundry.org/cf-operator/pkg/kube/apis/boshdeployment/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/names"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/versionedsecretstore"
)

// Resolver resolves references from bdpl CRD to a BOSH manifest
type Resolver struct {
	client               client.Client
	versionedSecretStore versionedsecretstore.VersionedSecretStore
	newInterpolatorFunc  func() Interpolator
}

// NewInterpolatorFunc returns a fresh Interpolator
type NewInterpolatorFunc func() Interpolator

// NewResolver constructs a resolver
func NewResolver(client client.Client, f NewInterpolatorFunc) *Resolver {
	return &Resolver{
		client:               client,
		newInterpolatorFunc:  f,
		versionedSecretStore: versionedsecretstore.NewVersionedSecretStore(client),
	}
}

// DesiredManifest reads the versioned secret created by the variable interpolation job
// and unmarshals it into a Manifest object
func (r *Resolver) DesiredManifest(ctx context.Context, boshDeploymentName, namespace string) (*Manifest, error) {
	// unversioned desired manifest name
	secretName := names.DesiredManifestName(boshDeploymentName, "")

	secret, err := r.versionedSecretStore.Latest(ctx, namespace, secretName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read versioned secret for desired manifest")
	}

	manifestData := secret.Data["manifest.yaml"]

	manifest, err := LoadYAML(manifestData)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to unmarshal manifest from secret '%s'", secretName)
	}

	return manifest, nil
}

// WithOpsManifest returns manifest referenced by our bdpl CRD
// The resulting manifest has variables interpolated and ops files applied.
// It is the 'with-ops' manifest.
func (r *Resolver) WithOpsManifest(instance *bdc.BOSHDeployment, namespace string) (*Manifest, error) {
	interpolator := r.newInterpolatorFunc()
	spec := instance.Spec
	manifest := &Manifest{}
	var (
		m   string
		err error
	)

	m, err = r.resourceData(namespace, spec.Manifest.Type, spec.Manifest.Ref, bdc.ManifestSpecName)
	if err != nil {
		return manifest, err
	}

	// Get the deployment name from the manifest
	manifest, err = LoadYAML([]byte(m))
	if err != nil {
		return manifest, errors.Wrapf(err, "failed to unmarshal manifest")
	}

	// Interpolate implicit variables
	vars := r.implicitVariables(manifest, m)
	for _, v := range vars {
		varData, err := r.resourceData(namespace, bdc.SecretType, names.CalculateSecretName(names.DeploymentSecretTypeImplicitVariable, instance.GetName(), v), bdc.ImplicitVariableKeyName)
		if err != nil {
			return manifest, errors.Wrapf(err, "failed to load secret for variable '%s'", v)
		}

		m = strings.Replace(m, fmt.Sprintf("((%s))", v), varData, -1)
	}

	// Interpolate manifest with ops if exist
	ops := spec.Ops
	if len(ops) == 0 {
		manifest, err := LoadYAML([]byte(m))
		return manifest, err
	}

	for _, op := range ops {
		opsData, err := r.resourceData(namespace, op.Type, op.Ref, bdc.OpsSpecName)
		if err != nil {
			return manifest, err
		}
		err = interpolator.BuildOps([]byte(opsData))
		if err != nil {
			return manifest, errors.Wrapf(err, "failed to build ops with: %#v", opsData)
		}
	}

	bytes, err := interpolator.Interpolate([]byte(m))
	if err != nil {
		return manifest, errors.Wrapf(err, "failed to interpolate %#v", m)
	}

	manifest, err = LoadYAML(bytes)

	return manifest, err
}

// resourceData resolves different manifest reference types and returns the resource's data
func (r *Resolver) resourceData(namespace string, resType string, name string, key string) (string, error) {
	var (
		data string
		ok   bool
	)

	switch resType {
	case bdc.ConfigMapType:
		opsConfig := &corev1.ConfigMap{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, opsConfig)
		if err != nil {
			return data, errors.Wrapf(err, "failed to retrieve %s from configmap '%s/%s' via client.Get", key, namespace, name)
		}
		data, ok = opsConfig.Data[key]
		if !ok {
			return data, fmt.Errorf("configMap '%s/%s' doesn't contain key %s", namespace, name, key)
		}
	case bdc.SecretType:
		opsSecret := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, opsSecret)
		if err != nil {
			return data, errors.Wrapf(err, "failed to retrieve %s from secret '%s/%s' via client.Get", key, namespace, name)
		}
		encodedData, ok := opsSecret.Data[key]
		if !ok {
			return data, fmt.Errorf("secret '%s/%s' doesn't contain key %s", namespace, name, key)
		}
		data = string(encodedData)
	case bdc.URLType:
		httpResponse, err := http.Get(name)
		if err != nil {
			return data, errors.Wrapf(err, "failed to resolve %s from url '%s' via http.Get", key, name)
		}
		body, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			return data, errors.Wrapf(err, "failed to read %s response body '%s' via ioutil", key, name)
		}
		data = string(body)
	default:
		return data, fmt.Errorf("unrecognized %s ref type %s", key, name)
	}

	return data, nil
}

// implicitVariables returns a list of all implicit variables in a manifest
func (r *Resolver) implicitVariables(m *Manifest, rawManifest string) []string {
	varMap := make(map[string]bool)

	// Collect all variables
	varRegexp := regexp.MustCompile(`\(\((!?[-/\.\w\pL]+)\)\)`)
	for _, match := range varRegexp.FindAllStringSubmatch(rawManifest, -1) {
		// Remove subfields from the match, e.g. ca.private_key -> ca
		fieldRegexp := regexp.MustCompile(`[^\.]+`)
		main := fieldRegexp.FindString(match[1])

		varMap[main] = true
	}

	// Remove the explicit ones
	for _, v := range m.Variables {
		varMap[v.Name] = false
	}

	names := []string{}
	for k, v := range varMap {
		if v {
			names = append(names, k)
		}
	}

	return names
}
