package promotion

import (
	"github.com/golang/glog"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register("PromoteHostnameAndSubdomainAnnotations", func(config io.Reader) (admission.Interface, error) {
		return NewPromoteHostnameAndSubdomainAnnotations(), nil
	})
}

type hostnameAndSubdomainPromotion struct {
	*admission.Handler
}

func (h *hostnameAndSubdomainPromotion) updateObject(obj runtime.Object) {
	pod, ok := obj.(*api.Pod)
	if !ok {
		return
	}
	if value, ok := pod.Annotations["pod.beta.kubernetes.io/hostname"]; ok && pod.Spec.Hostname == "" {
		glog.V(4).Infof("setting pod.Spec.Hostname to '%s'", value)
		pod.Spec.Hostname = value
	}
	if value, ok := pod.Annotations["pod.beta.kubernetes.io/subdomain"]; ok && pod.Spec.Subdomain == "" {
		glog.V(4).Infof("setting pod.Spec.Subdomain to '%s'", value)
		pod.Spec.Subdomain = value
	}
}

func (h *hostnameAndSubdomainPromotion) Admit(attributes admission.Attributes) error {
	// Ignore all calls to subresources or resources other than pods.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("pods") {
		return nil
	}
	if obj := attributes.GetObject(); obj != nil {
		h.updateObject(obj)
	}
	if obj := attributes.GetOldObject(); obj != nil {
		h.updateObject(obj)
	}
	return nil
}

func NewPromoteHostnameAndSubdomainAnnotations() admission.Interface {
	return &hostnameAndSubdomainPromotion{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}
