package operator

import (
	"fmt"
	opv1 "github.com/openshift/api/operator/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/openshift/api/config/v1"
	fakeconfig "github.com/openshift/client-go/config/clientset/versioned/fake"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
)

func TestWithCustomLabels(t *testing.T) {

	infraObj := &v1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: v1.InfrastructureStatus{
			InfrastructureName: "test-vbc3g",
			PlatformStatus: &v1.PlatformStatus{
				GCP: &v1.GCPPlatformStatus{
					ProjectID: "test",
					Region:    "test",
				},
			},
		},
	}

	tmplDeployObj := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "csi-driver",
							Image: "example.io/example-csi-driver",
							Args: []string{
								"--endpoint=$(CSI_ENDPOINT)",
								"--logtostderr",
								"--v=2",
								"--nodeid=2",
								"--controller=true",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/etc/cloud-sa/service_account.json",
								},
								{
									Name:  "CSI_ENDPOINT",
									Value: "unix:///var/lib/csi/sockets/pluginproxy/csi.sock",
								},
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "test",
										},
									},
								},
							},
						},
						{
							Name:  "test-driver",
							Image: "example.io/example-test-driver",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		labels        []v1.GCPResourceLabel
		expArgList    string
		createInfraCR bool
		wantErr       bool
	}{
		{
			name:          "labels not configured",
			labels:        []v1.GCPResourceLabel{},
			expArgList:    fmt.Sprintf("--extra-labels=%s", fmt.Sprintf(ocpDefaultLabelFmt, infraObj.Status.InfrastructureName)),
			createInfraCR: true,
			wantErr:       false,
		},
		{
			name: "labels configured",
			labels: []v1.GCPResourceLabel{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key:   "key2",
					Value: "value2",
				},
				{
					Key:   "key3",
					Value: "value3",
				},
			},
			expArgList: fmt.Sprintf("--extra-labels=key1=value1,key2=value2,"+
				"key3=value3,%s", fmt.Sprintf(ocpDefaultLabelFmt, infraObj.Status.InfrastructureName)),
			createInfraCR: true,
			wantErr:       false,
		},
		{
			name:          "Infrastructure CR does not exist",
			labels:        []v1.GCPResourceLabel{},
			expArgList:    "",
			createInfraCR: false,
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0)
			if test.createInfraCR {
				infraObj.Status.PlatformStatus.GCP.ResourceLabels = test.labels
				objs = append(objs, infraObj)
			}
			configClient := fakeconfig.NewSimpleClientset(objs...)
			configInformerFactory := configinformers.NewSharedInformerFactory(configClient, 0)
			if test.createInfraCR {
				configInformerFactory.Config().V1().Infrastructures().Informer().GetIndexer().Add(infraObj)
			}

			deployment := tmplDeployObj.DeepCopy()
			updDeployment := tmplDeployObj.DeepCopy()
			if test.expArgList != "" {
				updDeployment.Spec.Template.Spec.Containers[0].Args = append(
					updDeployment.Spec.Template.Spec.Containers[0].Args,
					test.expArgList,
				)
			}

			err := withCustomLabels(configInformerFactory.Config().V1().Infrastructures().Lister())(nil, deployment)
			if (err != nil) != test.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !equality.Semantic.DeepEqual(deployment, updDeployment) {
				t.Errorf("unexpected deployment want: %+v got: %+v", updDeployment, deployment)
			}
		})
	}
}

func TestWithCustomResourceTags(t *testing.T) {
	infraObj := &v1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: v1.InfrastructureStatus{
			InfrastructureName: "test-dfgh2",
			PlatformStatus: &v1.PlatformStatus{
				GCP: &v1.GCPPlatformStatus{
					ProjectID: "test",
					Region:    "test",
				},
			},
		},
	}

	tmplDeployObj := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "csi-driver",
							Image: "example.io/example-csi-driver",
							Args: []string{
								"--endpoint=$(CSI_ENDPOINT)",
								"--logtostderr",
								"--v=2",
								"--nodeid=2",
								"--controller=true",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/etc/cloud-sa/service_account.json",
								},
								{
									Name:  "CSI_ENDPOINT",
									Value: "unix:///var/lib/csi/sockets/pluginproxy/csi.sock",
								},
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "test",
										},
									},
								},
							},
						},
						{
							Name:  "test-driver",
							Image: "example.io/example-test-driver",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		tags          []v1.GCPResourceTag
		expArgList    string
		createInfraCR bool
		wantErr       bool
	}{
		{
			name:          "user tags not configured",
			tags:          []v1.GCPResourceTag{},
			expArgList:    "",
			createInfraCR: true,
			wantErr:       false,
		},
		{
			name: "user tags configured",
			tags: []v1.GCPResourceTag{
				{
					ParentID: "openshift",
					Key:      "key1",
					Value:    "value1",
				},
				{
					ParentID: "openshift",
					Key:      "key2",
					Value:    "value2",
				},
				{
					ParentID: "openshift",
					Key:      "key3",
					Value:    "value3",
				},
			},
			expArgList:    "--resource-tags=openshift/key1/value1,openshift/key2/value2,openshift/key3/value3",
			createInfraCR: true,
			wantErr:       false,
		},
		{
			name:          "Infrastructure CR does not exist",
			tags:          []v1.GCPResourceTag{},
			expArgList:    "",
			createInfraCR: false,
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0)
			if test.createInfraCR {
				infraObj.Status.PlatformStatus.GCP.ResourceTags = test.tags
				objs = append(objs, infraObj)
			}
			configClient := fakeconfig.NewSimpleClientset(objs...)
			configInformerFactory := configinformers.NewSharedInformerFactory(configClient, 0)
			if test.createInfraCR {
				configInformerFactory.Config().V1().Infrastructures().Informer().GetIndexer().Add(infraObj)
			}

			deployment := tmplDeployObj.DeepCopy()
			updDeployment := tmplDeployObj.DeepCopy()
			if test.expArgList != "" {
				updDeployment.Spec.Template.Spec.Containers[0].Args = append(
					updDeployment.Spec.Template.Spec.Containers[0].Args,
					test.expArgList,
				)
			}

			err := withCustomResourceTags(configInformerFactory.Config().V1().Infrastructures().Lister())(nil, deployment)
			if (err != nil) != test.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !equality.Semantic.DeepEqual(deployment, updDeployment) {
				t.Errorf("unexpected deployment want: %+v got: %+v", updDeployment, deployment)
			}
		})
	}
}

func TestWIFCredentialsRequestHook(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedFields map[string]interface{}
		wantErr        bool
		errMsg         string
	}{
		{
			name: "All required variables present",
			envVars: map[string]string{
				"PROJECT_NUMBER":        "123456",
				"POOL_ID":               "my-pool",
				"PROVIDER_ID":           "my-provider",
				"SERVICE_ACCOUNT_EMAIL": "sa@example.com",
			},
			expectedFields: map[string]interface{}{
				"spec.providerSpec.poolID":              "my-pool",
				"spec.providerSpec.providerID":          "my-provider",
				"spec.providerSpec.serviceAccountEmail": "sa@example.com",
				"spec.providerSpec.audience":            "//iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/my-pool/providers/my-provider",
			},
			wantErr: false,
		},
		{
			name:    "No variables present",
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "Missing some required variables",
			envVars: map[string]string{
				"PROJECT_NUMBER": "123456",
				"POOL_ID":        "my-pool",
			},
			wantErr: true,
			errMsg:  "cluster Workload Identity Federation environment detected, but some required environment variable(s) are missing: PROVIDER_ID, SERVICE_ACCOUNT_EMAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			spec := &opv1.OperatorSpec{}
			cr := &unstructured.Unstructured{
				Object: make(map[string]interface{}),
			}

			err := wifCredentialsRequestHook(spec, cr)

			if (err != nil) != tt.wantErr {
				t.Errorf("wifCredentialsRequestHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("wifCredentialsRequestHook() error message = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if !tt.wantErr && len(tt.expectedFields) > 0 {
				for path, expectedValue := range tt.expectedFields {
					parts := strings.Split(path, ".")
					value, exists, err := unstructured.NestedFieldNoCopy(cr.Object, parts...)
					if err != nil {
						t.Errorf("Error getting nested field %s: %v", path, err)
					}
					if !exists {
						t.Errorf("Expected field %s does not exist", path)
					}
					if value != expectedValue {
						t.Errorf("Field %s = %v, want %v", path, value, expectedValue)
					}
				}
			}
		})
	}
}
