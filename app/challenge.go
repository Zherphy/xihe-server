package app

import (
	"errors"
	"strings"

	"github.com/opensourceways/xihe-server/domain"
	"github.com/opensourceways/xihe-server/domain/challenge"
	"github.com/opensourceways/xihe-server/domain/repository"
	"github.com/opensourceways/xihe-server/utils"
)

type ChallengeService interface {
	Apply(*CompetitorApplyCmd) error
	GetCompetitor(domain.Account) (ChallengeCompetitorInfoDTO, error)
	GetAIQuestions(domain.Account) (AIQuestionDTO, error)
	SubmitAIQuestionAnswer(domain.Account, *AIQuestionAnswerSubmitCmd) (int, error)
}

type challengeService struct {
	comptitions []domain.CompetitionIndex
	aiQuestion  challenge.AIQuestionInfo
	delimiter   string

	competitionRepo repository.Competition
	aiQuestionRepo  repository.AIQuestion
	helper          challenge.Challenge
	encryption      utils.SymmetricEncryption
}

func NewChallengeService(
	competitionRepo repository.Competition,
	aiQuestionRepo repository.AIQuestion,
	helper challenge.Challenge,
	encryption utils.SymmetricEncryption,
) ChallengeService {
	v := helper.GetChallenge()

	s := &challengeService{
		competitionRepo: competitionRepo,
		aiQuestionRepo:  aiQuestionRepo,
		encryption:      encryption,
		helper:          helper,
		delimiter:       ",",
	}

	s.comptitions = make([]domain.CompetitionIndex, len(v.Competition))

	for i, cid := range v.Competition {
		s.comptitions[i] = domain.CompetitionIndex{
			Id:    cid,
			Phase: domain.CompetitionPhasePreliminary,
		}
	}

	s.aiQuestion = v.AIQuestionInfo

	return s
}

func (s *challengeService) Apply(cmd *CompetitorApplyCmd) error {
	c := cmd.toCompetitor()
	for i := range s.comptitions {
		// TODO allow re-apply
		err := s.competitionRepo.SaveCompetitor(&s.comptitions[i], c)
		if err != nil {
			return err
		}
	}

	// TODO allow re-apply
	return s.aiQuestionRepo.SaveCompetitor(s.aiQuestion.AIQuestionId, c)
}

func (s *challengeService) GetCompetitor(user domain.Account) (
	ChallengeCompetitorInfoDTO, error,
) {
	dto := ChallengeCompetitorInfoDTO{}

	for i := range s.comptitions {
		isCompetitor, score, err := s.getCompetitorOfCompetition(
			&s.comptitions[i], user,
		)

		if err != nil || !isCompetitor {
			return dto, err
		}

		dto.Score += score
	}

	isCompetitor, score, err := s.getCompetitorOfAIQuestion(s.aiQuestion.AIQuestionId, user)

	if err == nil && isCompetitor {
		dto.IsCompetitor = true
		dto.Score += score
	}

	return dto, err
}

func (s *challengeService) getCompetitorOfCompetition(
	index *domain.CompetitionIndex, user domain.Account,
) (isCompetitor bool, score int, err error) {

	isCompetitor, submissions, err := s.competitionRepo.GetCompetitorAndSubmission(
		index, user,
	)
	if err != nil || !isCompetitor {
		return
	}

	score = s.helper.CalcCompetitionScore(submissions)

	return
}

func (s *challengeService) getCompetitorOfAIQuestion(
	cid string, user domain.Account,
) (isCompetitor bool, score int, err error) {

	isCompetitor, scores, err := s.aiQuestionRepo.GetCompetitorAndScores(cid, user)
	if err != nil || !isCompetitor {
		return
	}

	for _, v := range scores {
		if v > score {
			score = v
		}
	}

	return
}

func (s *challengeService) SubmitAIQuestionAnswer(competitor domain.Account, cmd *AIQuestionAnswerSubmitCmd) (
	score int, err error,
) {
	now := utils.Now()

	v, err := s.aiQuestionRepo.GetSubmission(
		s.aiQuestion.AIQuestionId, competitor, utils.ToDate(now),
	)
	if err != nil {
		return
		// new
	}

	if v.Status != domain.AIQuestionStatusStart {
		err = errors.New("can't submit")

		return
	}

	if now > v.Expiry {
		err = errors.New("it is timeout")

		return
	}

	if cmd.Times != v.Times {
		err = errors.New("unmatched times")

		return
	}

	answer, err := s.decryptAnswer(cmd.Answer)
	if err != nil {
		return
	}

	if len(cmd.Result) != len(answer) {
		err = errors.New("invalid result")

		return
	}

	score = s.helper.CalcAIQuestionScore(cmd.Result, answer)
	if score > v.Score {
		v.Score = score
	}

	v.Status = domain.AIQuestionStatusEnd

	_, err = s.aiQuestionRepo.SaveSubmission(
		s.aiQuestion.AIQuestionId, &v,
	)

	return
}

func (s *challengeService) GetAIQuestions(competitor domain.Account) (dto AIQuestionDTO, err error) {
	now := utils.Now()
	date := utils.ToDate(now)
	expiry := now + int64((s.aiQuestion.Timeout+10)*60)

	v, err := s.aiQuestionRepo.GetSubmission(
		s.aiQuestion.AIQuestionId, competitor, date,
	)
	if err != nil {
		//return

		// gen question first to avoid occupying a times.
		if err = s.genAIQuestions(&dto); err != nil {
			return
		}

		// new
		v = domain.QuestionSubmission{
			Account: competitor,
			Date:    date,
			Status:  domain.AIQuestionStatusStart,
			Expiry:  expiry,
			Times:   1,
		}

		_, err = s.aiQuestionRepo.SaveSubmission(s.aiQuestion.AIQuestionId, &v)

		dto.Times = v.Times

		return
	}

	if v.Times >= s.aiQuestion.RetryTimes {
		err = errors.New("exceed max times")

		return
	}

	if v.Status == domain.AIQuestionStatusStart && now < v.Expiry {
		err = errors.New("it is in-progress")

		return
	}

	// gen question first to avoid occupying a times.
	if err = s.genAIQuestions(&dto); err != nil {
		return
	}

	v.Status = domain.AIQuestionStatusStart
	v.Expiry = expiry
	v.Times++

	_, err = s.aiQuestionRepo.SaveSubmission(s.aiQuestion.AIQuestionId, &v)

	dto.Times = v.Times

	return
}

func (s *challengeService) genAIQuestions(dto *AIQuestionDTO) (err error) {
	choice, completion := s.helper.GenAIQuestionNums()
	choices, completions, err := s.aiQuestionRepo.GetQuestions(
		s.aiQuestion.QuestionPoolId, choice, completion,
	)
	if err != nil {
		return
	}

	n := len(choice)
	answers := make([]string, n+len(completion))
	dto.Choices = make([]ChoiceQuestionDTO, n)

	for i := range choices {
		item := &choices[i]

		dto.Choices[i] = ChoiceQuestionDTO{
			Desc:    item.Desc,
			Options: item.Options,
		}

		answers[i] = item.Answer
	}

	dto.Completions = make([]string, len(completion))

	for i := range completions {
		item := &completions[i]

		dto.Completions[i] = item.Desc
		answers[i+n] = item.Answer
	}

	str, err := s.encryptAnswer(answers)
	if err == nil {
		dto.Answer = str
	}

	return
}

func (s *challengeService) encryptAnswer(answers []string) (string, error) {
	str := strings.Join(answers, s.delimiter)

	v, err := s.encryption.Encrypt([]byte(str))
	if err == nil {
		return string(v), nil
	}

	return "", err
}

func (s *challengeService) decryptAnswer(str string) ([]string, error) {
	v, err := s.encryption.Decrypt([]byte(str))
	if err != nil {
		return nil, err
	}

	return strings.Split(string(v), s.delimiter), nil
}