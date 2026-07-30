package operator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/library-go/pkg/crypto"
)

// WriteOperatorTLSConfig reads the cluster TLS profile from the APIServer CR
// and writes a GenericOperatorConfig file with TLS settings for the operator's
// own HTTPS endpoint (metrics, healthz).
// Returns empty string when running outside a cluster (local development).
// Returns an error when the APIServer CR cannot be read.
func WriteOperatorTLSConfig(ctx context.Context) (string, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.V(4).Infof("Not running in-cluster, skipping TLS profile injection: %v", err)
		return "", nil
	}

	client, err := configclient.NewForConfig(restConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create config client: %w", err)
	}

	apiServer, err := client.ConfigV1().APIServers().Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to read APIServer config: %w", err)
	}

	var profile *configv1.TLSSecurityProfile
	if crypto.ShouldHonorClusterTLSProfile(apiServer.Spec.TLSAdherence) {
		profile = apiServer.Spec.TLSSecurityProfile
		klog.Infof("TLS adherence policy is %q, using cluster TLS profile", apiServer.Spec.TLSAdherence)
	} else {
		klog.Infof("TLS adherence policy is %q, using default Intermediate profile", apiServer.Spec.TLSAdherence)
	}

	return writeConfig(resolveTLSProfile(profile))
}

func writeConfig(minTLSVersion string, cipherSuites []string) (string, error) {
	configDir := "/tmp/gcp-filestore-operator-tls"
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create TLS config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	content := buildOperatorConfig(minTLSVersion, cipherSuites)

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write operator TLS config: %w", err)
	}

	klog.Infof("Wrote operator TLS config to %s (minTLSVersion=%s, %d cipher suites)", configPath, minTLSVersion, len(cipherSuites))
	return configPath, nil
}

// resolveTLSProfile resolves a TLSSecurityProfile to concrete minTLSVersion and
// IANA cipher suite names. When profile is nil, defaults to Intermediate.
func resolveTLSProfile(profile *configv1.TLSSecurityProfile) (string, []string) {
	if profile == nil {
		spec := configv1.TLSProfiles[crypto.DefaultTLSProfileType]
		return string(spec.MinTLSVersion), crypto.OpenSSLToIANACipherSuites(spec.Ciphers)
	}

	var spec *configv1.TLSProfileSpec
	switch profile.Type {
	case configv1.TLSProfileCustomType:
		if profile.Custom != nil {
			spec = &profile.Custom.TLSProfileSpec
		}
	case configv1.TLSProfileOldType, configv1.TLSProfileIntermediateType, configv1.TLSProfileModernType:
		spec = configv1.TLSProfiles[profile.Type]
	}

	if spec == nil {
		spec = configv1.TLSProfiles[crypto.DefaultTLSProfileType]
	}

	return string(spec.MinTLSVersion), crypto.OpenSSLToIANACipherSuites(spec.Ciphers)
}

func buildOperatorConfig(minTLSVersion string, cipherSuites []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: operator.openshift.io/v1alpha1
kind: GenericOperatorConfig
servingInfo:
  minTLSVersion: %s
`, minTLSVersion)

	if len(cipherSuites) > 0 {
		b.WriteString("  cipherSuites:\n")
		for _, cs := range cipherSuites {
			fmt.Fprintf(&b, "  - %s\n", cs)
		}
	}

	return b.String()
}
