package repository

import (
	"github.com/opensourceways/xihe-server/domain"
)

type AIQuestion interface {
	GetResult(string) ([]domain.QuestionSubmissionInfo, error)

	GetCompetitorAndScores(string, domain.Account) (bool, []int, error)

	SaveCompetitor(string, *domain.CompetitorInfo) error

	GetQuestions(pool string, choice, completion []int) (
		[]domain.ChoiceQuestion, []domain.CompletionQuestion, error,
	)

	GetSubmission(qid string, user domain.Account, date string) (
		domain.QuestionSubmission, error,
	)

	SaveSubmission(qid string, v *domain.QuestionSubmission) (string, error)
}
