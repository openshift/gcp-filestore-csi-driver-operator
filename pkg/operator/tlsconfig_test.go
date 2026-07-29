package operator

import (
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
)

func TestResolveTLSProfile(t *testing.T) {
	tests := []struct {
		name            string
		profile         *configv1.TLSSecurityProfile
		wantMinVersion  string
		wantCipherCount int
		wantFirstCipher string
	}{
		{
			name:           "nil profile defaults to Intermediate",
			profile:        nil,
			wantMinVersion: string(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].MinTLSVersion),
		},
		{
			name: "Old profile",
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileOldType,
			},
			wantMinVersion: string(configv1.TLSProfiles[configv1.TLSProfileOldType].MinTLSVersion),
		},
		{
			name: "Intermediate profile",
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileIntermediateType,
			},
			wantMinVersion: string(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].MinTLSVersion),
		},
		{
			name: "Modern profile",
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileModernType,
			},
			wantMinVersion: string(configv1.TLSProfiles[configv1.TLSProfileModernType].MinTLSVersion),
		},
		{
			name: "Custom profile with spec",
			profile: &configv1.TLSSecurityProfile{
				Type: configv1.TLSProfileCustomType,
				Custom: &configv1.CustomTLSProfile{
					TLSProfileSpec: configv1.TLSProfileSpec{
						MinTLSVersion: configv1.VersionTLS12,
						Ciphers:       []string{"ECDHE-RSA-AES128-GCM-SHA256"},
					},
				},
			},
			wantMinVersion:  string(configv1.VersionTLS12),
			wantCipherCount: 1,
			wantFirstCipher: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		},
		{
			name: "Custom profile with nil Custom spec falls back to Intermediate",
			profile: &configv1.TLSSecurityProfile{
				Type:   configv1.TLSProfileCustomType,
				Custom: nil,
			},
			wantMinVersion: string(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].MinTLSVersion),
		},
		{
			name: "Unknown profile type falls back to Intermediate",
			profile: &configv1.TLSSecurityProfile{
				Type: "UnknownType",
			},
			wantMinVersion: string(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].MinTLSVersion),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minVersion, ciphers := resolveTLSProfile(tt.profile)

			if minVersion != tt.wantMinVersion {
				t.Errorf("minTLSVersion = %q, want %q", minVersion, tt.wantMinVersion)
			}

			if tt.wantCipherCount > 0 && len(ciphers) != tt.wantCipherCount {
				t.Errorf("cipher count = %d, want %d", len(ciphers), tt.wantCipherCount)
			}

			if tt.wantFirstCipher != "" && (len(ciphers) == 0 || ciphers[0] != tt.wantFirstCipher) {
				got := ""
				if len(ciphers) > 0 {
					got = ciphers[0]
				}
				t.Errorf("first cipher = %q, want %q", got, tt.wantFirstCipher)
			}

			if tt.wantCipherCount == 0 && len(ciphers) == 0 {
				t.Error("expected non-empty cipher list for standard profiles")
			}
		})
	}
}

func TestBuildOperatorConfig(t *testing.T) {
	tests := []struct {
		name         string
		minVersion   string
		cipherSuites []string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "standard config with ciphers",
			minVersion:   "VersionTLS12",
			cipherSuites: []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"},
			wantContains: []string{
				"apiVersion: operator.openshift.io/v1alpha1",
				"kind: GenericOperatorConfig",
				"minTLSVersion: VersionTLS12",
				"cipherSuites:",
				"  - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				"  - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			},
		},
		{
			name:         "config without ciphers",
			minVersion:   "VersionTLS13",
			cipherSuites: nil,
			wantContains: []string{
				"minTLSVersion: VersionTLS13",
			},
			wantAbsent: []string{
				"cipherSuites:",
			},
		},
		{
			name:         "empty cipher slice",
			minVersion:   "VersionTLS12",
			cipherSuites: []string{},
			wantAbsent: []string{
				"cipherSuites:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOperatorConfig(tt.minVersion, tt.cipherSuites)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("output missing expected content %q\ngot:\n%s", want, result)
				}
			}

			for _, absent := range tt.wantAbsent {
				if strings.Contains(result, absent) {
					t.Errorf("output should not contain %q\ngot:\n%s", absent, result)
				}
			}
		})
	}
}
