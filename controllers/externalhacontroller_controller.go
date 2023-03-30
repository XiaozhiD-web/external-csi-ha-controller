/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	csipluginsv1 "github.com/external-csi-ha-controller/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logf.Log.WithName("controller_ExternalHaController")

// ExternalHaControllerReconciler reconciles a ExternalHaController object
type ExternalHaControllerReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=csiplugins.spdbdev.io,resources=externalhacontrollers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiplugins.spdbdev.io,resources=externalhacontrollers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiplugins.spdbdev.io,resources=externalhacontrollers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExternalHaController object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ExternalHaControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling ExternalHaController")

	externalHaController := &csipluginsv1.ExternalHaController{}
	err := r.Client.Get(ctx, req.NamespacedName, externalHaController)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("haController resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Info("Failed to get externalHaController.")
		return reconcile.Result{}, err
	}

	pvcList := v1.PersistentVolumeClaimList{}
	err = r.Client.List(ctx, &pvcList)
	if err != nil {
		reqLogger.Error(err, "Failed to list PersistentVolumeClaim", pvcList)

	}

	var pvcs []string
	var pvs []string
	if len(pvcList.Items) != 0 {
		for _, pvc := range pvcList.Items {
			if pvc.Annotations["volume.kubernetes.io/storage-provisioner"] == "hostpath.csi.k8s.io" {
				pvcs = append(pvcs, pvc.Name)
				pvs = append(pvs, "pvc-"+string(pvc.UID))
			}

		}
	}

	vaList := storagev1.VolumeAttachmentList{}
	err = r.Client.List(ctx, &vaList)
	if err != nil {
		reqLogger.Error(err, "Failed to list VolumeAttachment")
	}

	podList := v1.PodList{}
	err = r.Client.List(ctx, &podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pod")
	}

	if externalHaController.Spec.DeletePod == true {
		for _, pod := range podList.Items {
			if _, exist := pod.Labels["deleteNow"]; exist {
				err = r.Client.Delete(ctx, &pod, &client.DeleteOptions{
					Preconditions: v12.NewUIDPreconditions(string(pod.UID)),
				})
				if err != nil {
					reqLogger.Error(err, "Failed to delete pod")
				}
			}

			for _, volume := range pod.Spec.Volumes {
				if volume.PersistentVolumeClaim != nil {
					for _, pvc := range pvcs {
						if volume.PersistentVolumeClaim.ClaimName == pvc {
							err = r.Client.Delete(ctx, &pod, &client.DeleteOptions{
								Preconditions: v12.NewUIDPreconditions(string(pod.UID)),
							})
							if err != nil {
								reqLogger.Error(err, "Failed to delete pod")
							}
						}
					}
				}
			}
		}
		for _, pv := range pvs {
			for _, va := range vaList.Items {
				if *va.Spec.Source.PersistentVolumeName == pv || va.Spec.Attacher == "driver.longhorn.io" {
					err = r.Client.Delete(ctx, &va, &client.DeleteOptions{
						Preconditions: v12.NewUIDPreconditions(string(va.UID)),
					})
					if err != nil {
						reqLogger.Error(err, "Failed to delete va")
					}
				}
			}
		}

	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExternalHaControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csipluginsv1.ExternalHaController{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
