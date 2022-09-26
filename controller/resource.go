package controller

import (
	"errors"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/opensourceways/xihe-server/app"
	"github.com/opensourceways/xihe-server/domain"
)

type relatedResourceAddRequest struct {
	Owner string `json:"owner" required:"true"`
	Name  string `json:"name" required:"true"`
}

func (req *relatedResourceAddRequest) toModelCmd() (
	owner domain.Account, name domain.ModelName, err error,
) {
	if owner, err = domain.NewAccount(req.Owner); err != nil {
		return
	}

	name, err = domain.NewModelName(req.Name)

	return
}

func (req *relatedResourceAddRequest) toDatasetCmd() (
	owner domain.Account, name domain.DatasetName, err error,
) {
	if owner, err = domain.NewAccount(req.Owner); err != nil {
		return
	}

	name, err = domain.NewDatasetName(req.Name)

	return
}

type relatedResourceRemoveRequest struct {
	Owner string `json:"owner" required:"true"`
	Id    string `json:"id" required:"true"`
}

func (req *relatedResourceRemoveRequest) toCmd() (
	cmd domain.ResourceIndex, err error,
) {
	if cmd.Owner, err = domain.NewAccount(req.Owner); err != nil {
		return
	}

	if req.Id == "" {
		err = errors.New("missing id")
	}

	return
}

func convertToRelatedResource(data interface{}) (r app.ResourceDTO) {
	switch data.(type) {
	case domain.Model:
		v := data.(domain.Model)
		r.Id = v.Id
		r.Owner.Name = v.Owner.Account()
		//r.Owner.AvatarId =

		r.Name = v.Name.ResourceName()
		r.Type = domain.ResourceModel
		//r.UpdateAt
		r.LikeCount = v.LikeCount
		//r.DownloadCount =

	case domain.Dataset:
		v := data.(domain.Dataset)
		r.Id = v.Id
		r.Owner.Name = v.Owner.Account()
		//r.Owner.AvatarId =

		r.Name = v.Name.ResourceName()
		r.Type = domain.ResourceDataset
		//r.UpdateAt
		r.LikeCount = v.LikeCount
		//r.DownloadCount =
	}

	return
}

type resourceTagsUpdateRequest struct {
	ToAdd    []string `json:"add"`
	ToRemove []string `json:"remove"`
}

func (req *resourceTagsUpdateRequest) toCmd(
	validTags []domain.DomainTags,
) (cmd app.ResourceTagsUpdateCmd, err error) {

	err = errors.New("invalid cmd")

	if len(req.ToAdd) > 0 && len(req.ToRemove) > 0 {
		if sets.NewString(req.ToAdd...).HasAny(req.ToRemove...) {
			return
		}
	}

	tags := sets.NewString()

	for i := range validTags {
		for _, item := range validTags[i].Items {
			tags.Insert(item.Items...)
		}
	}

	if len(req.ToAdd) > 0 && !tags.HasAll(req.ToAdd...) {
		return
	}

	if len(req.ToRemove) > 0 && !tags.HasAll(req.ToRemove...) {
		return
	}

	cmd.ToAdd = req.ToAdd
	cmd.ToRemove = req.ToRemove
	err = nil

	return
}