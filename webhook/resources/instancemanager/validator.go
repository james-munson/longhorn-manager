package instancemanager

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	admissionregv1 "k8s.io/api/admissionregistration/v1"

	"github.com/longhorn/longhorn-manager/datastore"
	"github.com/longhorn/longhorn-manager/webhook/admission"

	longhorn "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	werror "github.com/longhorn/longhorn-manager/webhook/error"
)

type instanceManagerValidator struct {
	admission.DefaultValidator
	ds *datastore.DataStore
}

func NewValidator(ds *datastore.DataStore) admission.Validator {
	return &instanceManagerValidator{ds: ds}
}

func (i *instanceManagerValidator) Resource() admission.Resource {
	return admission.Resource{
		Name:       "instancemanagers",
		Scope:      admissionregv1.NamespacedScope,
		APIGroup:   longhorn.SchemeGroupVersion.Group,
		APIVersion: longhorn.SchemeGroupVersion.Version,
		ObjectType: &longhorn.InstanceManager{},
		OperationTypes: []admissionregv1.OperationType{
			admissionregv1.Create,
			admissionregv1.Update,
		},
	}
}

func (i *instanceManagerValidator) Create(request *admission.Request, newObj runtime.Object) error {
	im, ok := newObj.(*longhorn.InstanceManager)
	if !ok {
		return werror.NewInvalidError(fmt.Sprintf("%v is not a *longhorn.InstanceManager", newObj), "")
	}
	if err := i.validate(im); err != nil {
		return werror.NewInvalidError(err.Error(), "")
	}

	return nil
}

func (i *instanceManagerValidator) Update(request *admission.Request, oldObj runtime.Object, newObj runtime.Object) error {
	newIm, ok := newObj.(*longhorn.InstanceManager)
	if !ok {
		return werror.NewInvalidError(fmt.Sprintf("%v is not a *longhorn.InstanceManager", newObj), "")
	}

	if err := i.validate(newIm); err != nil {
		return werror.NewInvalidError(err.Error(), "")
	}

	return nil
}

func (i *instanceManagerValidator) validate(im *longhorn.InstanceManager) error {
	if im.Labels == nil {
		return fmt.Errorf("labels for instanceManager %s is not set", im.Name)
	}

	if im.OwnerReferences == nil {
		return fmt.Errorf("ownerReferences for instanceManager %s is not set", im.Name)
	}

	if im.Spec.Type == "" {
		return fmt.Errorf("type for instanceManager %s is not set", im.Name)
	}

	if im.Spec.DataEngine == "" {
		return fmt.Errorf("data engine for instanceManager %s is not set", im.Name)
	}

	if im.Spec.DataEngineSpec.V2.CPUMask != "" {
		err := i.ds.ValidateCPUMask(im.Spec.DataEngineSpec.V2.CPUMask)
		if err != nil {
			return werror.NewInvalidError(err.Error(), "")
		}
	}

	return nil
}
