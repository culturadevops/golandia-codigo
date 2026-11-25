// El recurso propio del operador (vídeo 517 de golandia). Reemplaza al que genera `kubebuilder create api`.
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PaginaWebSpec es lo que pide el usuario: cuántas réplicas y qué mensaje sirve la página.
type PaginaWebSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:MinLength=1
	Mensaje string `json:"mensaje"`
}

// PaginaWebStatus es lo que el operador observa y escribe.
type PaginaWebStatus struct {
	ReplicasListas int32 `json:"replicasListas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Listas",type=integer,JSONPath=`.status.replicasListas`
// +kubebuilder:printcolumn:name="Mensaje",type=string,JSONPath=`.spec.mensaje`

// PaginaWeb es una página estática servida por nginx, descrita en una sola línea de YAML.
type PaginaWeb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PaginaWebSpec   `json:"spec,omitempty"`
	Status PaginaWebStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PaginaWebList contiene una lista de PaginaWeb.
type PaginaWebList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PaginaWeb `json:"items"`
}

// kubebuilder 4.16 registra los tipos así (con runtime.SchemeBuilder; las versiones anteriores pasaban los tipos directamente).
func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PaginaWeb{}, &PaginaWebList{})
		return nil
	})
}
