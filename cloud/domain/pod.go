package domain

import (
	types "github.com/opensourceways/xihe-server/common/domain"
	otypes "github.com/opensourceways/xihe-server/domain"
	"github.com/opensourceways/xihe-server/utils"
)

type Pod struct {
	Id      string
	CloudId string
	Owner   otypes.Account
	Image   string
}

type PodInfo struct {
	Pod

	Status    PodStatus
	Expiry    PodExpiry
	Error     PodError
	AccessURL AccessURL
	CreatedAt types.Time
	CardsNum  CloudSpecCardsNum
}

func (r *Pod) IsOwner(owner otypes.Account) bool {
	return r.Owner == owner
}

func (p *PodInfo) CanRelease() bool {
	return p.Status.IsRunning() && !p.IsExpired()
}

func (p *PodInfo) IsExpired() bool {
	return utils.Now() > p.Expiry.PodExpiry()
}

func (p *PodInfo) IsFailedOrTerminated() bool {
	return p.Status.IsFailed() || p.IsTerminated()
}

func (p *PodInfo) IsHoldingAndNotExpired() bool {
	if p.IsExpired() {
		return false
	}

	return p.Status.IsCreating() || p.Status.IsStarting() || p.Status.IsRunning() || p.IsTerminating()
}

func (p *PodInfo) CheckGoodAndSet() bool {
	if !p.Error.IsGood() {
		p.Status, _ = NewPodStatus(CloudPodStatusFailed)
		return false
	}

	return true
}

func (p *PodInfo) StatusSetCreating() {
	p.Status, _ = NewPodStatus(CloudPodStatusCreating)
}

func (p *PodInfo) StatusSetRunning() {
	p.Status, _ = NewPodStatus(CloudPodStatusRunning)
}

func (p *PodInfo) StatusSetFailed() {
	p.Status, _ = NewPodStatus(CloudPodStatusFailed)
}

func (p *PodInfo) StatusSetTerminating() {
	p.Status, _ = NewPodStatus(CloudPodStatusTerminating)
}

func (p *PodInfo) StatusSetTerminated() {
	p.Status, _ = NewPodStatus(CloudPodStatusTerminated)
}

func (p *PodInfo) SetStatus() {
	if p.AccessURL.AccessURL() != "" {
		p.StatusSetRunning()
	} else {
		p.StatusSetFailed()
	}
}

func (p *PodInfo) SetDefaultExpiry() (err error) {
	if p.Expiry, err = NewPodExpiry(utils.Now() + 2*60*60); err != nil { // TODO config
		return
	}

	return
}

func (p *PodInfo) SetStartingPodInfo(
	cid string, owner otypes.Account, image ICloudImage, cardsNum CloudSpecCardsNum,
) (err error) {
	p.CloudId = cid
	p.Owner = owner
	p.Image = image.Image()
	p.CardsNum = cardsNum

	if p.Status, err = NewPodStatus(CloudPodStatusStarting); err != nil {
		return
	}

	return
}

func (p *PodInfo) GetCloudType() string {
	if p.CloudId == cloudIdCPU {
		return cloudTypeCPU
	}

	return cloudTypeNPU
}

func (p *PodInfo) IsCpu() bool {
	return p.CloudId == cloudIdCPU
}

func (p *PodInfo) IsAscend() bool {
	return p.CloudId == cloudIdNPU
}

func (p *PodInfo) IsTerminated() bool {
	return p.IsExpired()
}

func (p *PodInfo) IsTerminating() bool {
	return !p.IsExpired() && (p.Status.IsTerminated() || p.Status.IsTerminating())
}
