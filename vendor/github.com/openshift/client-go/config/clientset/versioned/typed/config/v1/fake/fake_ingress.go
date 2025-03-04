// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/openshift/api/config/v1"
	configv1 "github.com/openshift/client-go/config/applyconfigurations/config/v1"
	typedconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	gentype "k8s.io/client-go/gentype"
)

// fakeIngresses implements IngressInterface
type fakeIngresses struct {
	*gentype.FakeClientWithListAndApply[*v1.Ingress, *v1.IngressList, *configv1.IngressApplyConfiguration]
	Fake *FakeConfigV1
}

func newFakeIngresses(fake *FakeConfigV1) typedconfigv1.IngressInterface {
	return &fakeIngresses{
		gentype.NewFakeClientWithListAndApply[*v1.Ingress, *v1.IngressList, *configv1.IngressApplyConfiguration](
			fake.Fake,
			"",
			v1.SchemeGroupVersion.WithResource("ingresses"),
			v1.SchemeGroupVersion.WithKind("Ingress"),
			func() *v1.Ingress { return &v1.Ingress{} },
			func() *v1.IngressList { return &v1.IngressList{} },
			func(dst, src *v1.IngressList) { dst.ListMeta = src.ListMeta },
			func(list *v1.IngressList) []*v1.Ingress { return gentype.ToPointerSlice(list.Items) },
			func(list *v1.IngressList, items []*v1.Ingress) { list.Items = gentype.FromPointerSlice(items) },
		),
		fake,
	}
}
