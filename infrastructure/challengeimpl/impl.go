package challengeimpl

import (
	"math/rand"
	"time"

	"github.com/opensourceways/xihe-server/domain"
	"github.com/opensourceways/xihe-server/domain/challenge"
)

func NewChallenge(cfg *Config) challenge.Challenge {
	return &challengeImpl{*cfg}
}

type challengeImpl struct {
	cfg Config
}

func (impl *challengeImpl) GetChallenge() challenge.ChallengeInfo {
	info := &impl.cfg.AIQuestion

	return challenge.ChallengeInfo{
		Competition: impl.cfg.Competitions,

		AIQuestionInfo: challenge.AIQuestionInfo{
			AIQuestionId:   info.AIQuestionId,
			QuestionPoolId: info.QuestionPoolId,
			Timeout:        info.Timeout,
			RetryTimes:     info.RetryTimes,
		},
	}
}

func (impl *challengeImpl) CalcCompetitionScore(
	submissions []domain.CompetitionSubmissionInfo,
) int {
	for i := range submissions {
		if submissions[i].Status == impl.cfg.CompetitionSuccessStatus {
			return impl.cfg.CompetitionSuccessScore
		}
	}

	return 0
}

func (impl *challengeImpl) GenAIQuestionNums() (choice, completion []int) {
	cfg := impl.cfg.AIQuestion

	choice = impl.genRandoms(cfg.ChoiceQuestionsCount, cfg.ChoiceQuestionsNum)
	completion = impl.genRandoms(cfg.CompletionQuestionsCount, cfg.CompletionQuestionsNum)

	return
}

func (impl *challengeImpl) CalcAIQuestionScore(result, answer []string) (score int) {
	cfg := impl.cfg.AIQuestion

	num := cfg.ChoiceQuestionsNum
	for i := 0; i < num; i++ {
		if result[i] == answer[i] {
			score += cfg.ChoiceQuestionsScore
		}
	}

	total := num + cfg.CompletionQuestionsNum
	for i := num; i < total; i++ {
		if result[i] == answer[i] {
			score += cfg.CompletionQuestionsScore
		}
	}

	return
}

func (impl *challengeImpl) genRandoms(max, total int) []int {
	// set seed
	rand.Seed(time.Now().UnixNano())

	min := 1
	v := max - min
	i := 0
	m := make(map[int]struct{})
	r := make([]int, total)
	for {
		n := rand.Intn(v) + min

		if _, ok := m[n]; !ok {
			m[n] = struct{}{}
			r[i] = n
			if i++; i == total {
				break
			}
		}
	}

	return r
}