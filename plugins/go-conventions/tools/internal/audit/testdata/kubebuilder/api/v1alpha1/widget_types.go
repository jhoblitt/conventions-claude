package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type WidgetSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`
}

type Widget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WidgetSpec `json:"spec,omitempty"`
}

func (w *Widget) DeepCopyObject() any { return nil }
