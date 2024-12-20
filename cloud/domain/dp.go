package domain

import (
	"errors"
	"net/url"

	"github.com/opensourceways/xihe-server/utils"
)

const (
	CloudPodStatusStarting    = "starting"
	CloudPodStatusCreating    = "creating"
	CloudPodStatusFailed      = "failed"
	CloudPodStatusRunning     = "running"
	CloudPodStatusTerminated  = "terminated"
	CloudPodStatusTerminating = "terminating"
)

var cloudSpecCardsNumRange = map[int]struct{}{
	1: {},
	2: {},
	4: {},
	8: {},
}

// CloudName
type CloudName interface {
	CloudName() string
}

func NewCloudName(v string) (CloudName, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return cloudName(v), nil
}

type cloudName string

func (r cloudName) CloudName() string {
	return string(r)
}

// CloudSpec
type CloudSpecDesc interface {
	CloudSpecDesc() string
}

func NewCloudSpecDesc(v string) (CloudSpecDesc, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return cloudSpecDesc(v), nil
}

type cloudSpecDesc string

func (r cloudSpecDesc) CloudSpecDesc() string {
	return string(r)
}

// CloudSpecCardsNum
type CloudSpecCardsNum interface {
	CloudSpecCardsNum() int
}

func NewCloudSpecCardsNum(v int) (CloudSpecCardsNum, error) {
	if _, ok := cloudSpecCardsNumRange[v]; !ok {
		return nil, errors.New("the number of cards of npu is not supported")
	}

	return cloudSpecCardsNum(v), nil
}

type cloudSpecCardsNum int

func (r cloudSpecCardsNum) CloudSpecCardsNum() int {
	return int(r)
}

// CloudFeature
type CloudFeature interface {
	CloudFeature() string
}

func NewCloudFeature(v string) (CloudFeature, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return cloudFeature(v), nil
}

type cloudFeature string

func (r cloudFeature) CloudFeature() string {
	return string(r)
}

// CloudProcessor
type CloudProcessor interface {
	CloudProcessor() string
}

func NewCloudProcessor(v string) (CloudProcessor, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return cloudProcessor(v), nil
}

type cloudProcessor string

func (r cloudProcessor) CloudProcessor() string {
	return string(r)
}

// Credit
type Credit interface {
	Credit() int64
}

func NewCredit(v int64) (Credit, error) {
	if v < 0 {
		return nil, errors.New("invalid value")
	}

	return credit(v), nil
}

type credit int64

func (r credit) Credit() int64 {
	return int64(r)
}

// CloudLimited
type CloudLimited interface {
	CloudLimited() int
}

func NewCloudLimited(v int) (CloudLimited, error) {
	if v < 0 {
		return nil, errors.New("invalid value")
	}

	return cloudLimited(v), nil
}

type cloudLimited int

func (r cloudLimited) CloudLimited() int {
	return int(r)
}

// CloudRemain
type CloudRemain interface {
	CloudRemain() int
}

func NewCloudRemain(v int) (CloudRemain, error) {
	if v < 0 {
		return nil, errors.New("invalid value")
	}

	return cloudRemain(v), nil
}

type cloudRemain int

func (r cloudRemain) CloudRemain() int {
	return int(r)
}

// PodStatus
type PodStatus interface {
	PodStatus() string
	IsStarting() bool
	IsCreating() bool
	IsFailed() bool
	IsRunning() bool
	IsTerminated() bool
	IsTerminating() bool
}

func NewPodStatus(v string) (PodStatus, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return podStatus(v), nil
}

type podStatus string

func (r podStatus) PodStatus() string {
	return string(r)
}

func (r podStatus) IsStarting() bool {
	return r.PodStatus() == CloudPodStatusStarting
}

func (r podStatus) IsCreating() bool {
	return r.PodStatus() == CloudPodStatusCreating
}

func (r podStatus) IsFailed() bool {
	return r.PodStatus() == CloudPodStatusFailed
}

func (r podStatus) IsRunning() bool {
	return r.PodStatus() == CloudPodStatusRunning
}

func (r podStatus) IsTerminated() bool {
	return r.PodStatus() == CloudPodStatusTerminated
}

func (r podStatus) IsTerminating() bool {
	return r.PodStatus() == CloudPodStatusTerminating
}

// PodExpiry
type PodExpiry interface {
	PodExpiry() int64
	PodExpiryDate() string
}

func NewPodExpiry(v int64) (PodExpiry, error) {
	return podExpiry(v), nil
}

type podExpiry int64

func (r podExpiry) PodExpiry() int64 {
	return int64(r)
}

func (r podExpiry) PodExpiryDate() string {
	return utils.ToDate(r.PodExpiry())
}

// PodError
type PodError interface {
	PodError() string
	IsGood() bool
}

func NewPodError(v string) (PodError, error) {
	return podError(v), nil
}

type podError string

func (r podError) PodError() string {
	return string(r)
}

func (p podError) IsGood() bool {
	return p.PodError() == ""
}

// AccessURL
type AccessURL interface {
	AccessURL() string
}

func NewAccessURL(v string) (AccessURL, error) {
	if _, err := url.Parse(v); err != nil {
		return nil, errors.New("invalid url")
	}

	return accessURL(v), nil
}

type accessURL string

func (r accessURL) AccessURL() string {
	return string(r)
}

type CloudImageAlias interface {
	CloudImageAlias() string
}

func NewCloudImageAlias(v string) (CloudImageAlias, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return cloudImageAlias(v), nil
}

type cloudImageAlias string

func (a cloudImageAlias) CloudImageAlias() string {
	return string(a)
}

type ICloudImage interface {
	Image() string
}

func NewICloudImage(v string) (ICloudImage, error) {
	if v == "" {
		return nil, errors.New("empty value")
	}

	return cloudImage(v), nil
}

type cloudImage string

func (i cloudImage) Image() string {
	return string(i)
}
