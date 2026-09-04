package controller

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	widgetv1alpha1 "example.com/operator/api/v1alpha1"
)

type Reconciler struct{ ctrl.Manager }

func SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&widgetv1alpha1.Widget{}).Complete(&Reconciler{Manager: mgr})
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.FromContext(ctx).Info("reconciling", "name", req.Name)

	return ctrl.Result{}, nil
}
