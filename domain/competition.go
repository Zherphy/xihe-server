package domain

type CompetitionSummary struct {
	Id         string
	Name       CompetitionName
	Desc       CompetitionDesc
	Host       CompetitionHost
	Bonus      CompetitionBonus
	Status     CompetitionStatus
	Duration   CompetitionDuration
	Poster     URL
	ScoreOrder CompetitionScoreOrder
}

type Competition struct {
	CompetitionSummary

	Doc        URL
	Forum      Forum
	DatasetDoc URL
	DatasetURL URL
	Winners    Winners

	Type    CompetitionType
	Phase   CompetitionPhase
	Enabled bool
}

type CompetitorInfo struct {
	Account  Account
	Name     CompetitorName
	City     City
	Email    Email
	Phone    Phone
	Identity CompetitionIdentity
	Province Province
	Detail   map[string]string
}

type Competitor struct {
	CompetitorInfo

	Team     CompetitionTeam
	TeamRole TeamRole
}

type CompetitorSummary struct {
	IsCompetitor bool
	TeamId       string
	TeamRole     TeamRole
}

type CompetitionTeam struct {
	Id   string
	Name TeamName
}

type CompetitionRepo struct {
	TeamId     string
	Individual Account

	Owner Account
	Repo  ResourceName
}

type CompetitionSubmissionInfo struct {
	Id     string
	Status string
	Score  float32
}

func (info *CompetitionSubmissionInfo) IsSuccess() bool {
	return info.Status == competitionSubmissionStatusSuccess
}

type CompetitionSubmission struct {
	Id string

	TeamId     string
	Individual Account

	SubmitAt int64
	OBSPath  string
	Status   string
	Score    float32
}

func (info *CompetitionSubmission) IsSuccess() bool {
	return info.Status == competitionSubmissionStatusSuccess
}

func (r *CompetitionSubmission) IsTeamWork() bool {
	return r.TeamId != ""
}

func (r *CompetitionSubmission) Key() string {
	if r.TeamId != "" {
		return r.TeamId
	}

	return r.Individual.Account()
}

type CompetitionIndex struct {
	Id    string
	Phase CompetitionPhase
}

type CompetitionScoreOrder interface {
	IsBetterThanB(a, b float32) bool
}

func NewCompetitionScoreOrder(b bool) CompetitionScoreOrder {
	return smallerIsBetter(b)
}

type smallerIsBetter bool

func (order smallerIsBetter) IsBetterThanB(a, b float32) bool {
	if order {
		return a <= b
	}

	return a >= b
}
