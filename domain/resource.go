package domain

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

func ResourceTypeByName(n string) (ResourceType, error) {
	if strings.HasPrefix(n, ResourceProject) {
		return ResourceTypeProject, nil
	}

	if strings.HasPrefix(n, ResourceDataset) {
		return ResourceTypeDataset, nil
	}

	if strings.HasPrefix(n, ResourceModel) {
		return ResourceTypeModel, nil
	}

	return nil, errors.New("unknow resource")
}

// ResourceObject
type ResourceObject struct {
	Type ResourceType

	ResourceIndex
}

func (r *ResourceObject) String() string {
	return fmt.Sprintf(
		"%s_%s_%s",
		r.Owner.Account(),
		r.Type.ResourceType(),
		r.Id,
	)
}

type ResourceIndex struct {
	Owner Account
	Id    string
}

type RelatedResources []ResourceIndex

func (r RelatedResources) Has(index *ResourceIndex) bool {
	v := sets.NewString()

	for i := range ([]ResourceIndex)(r) {
		v.Insert(
			r[i].Owner.Account() + r[i].Id,
		)
	}

	return v.Has(index.Owner.Account() + index.Id)
}

func (r RelatedResources) Count() int {
	return len(r)
}