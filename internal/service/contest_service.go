package service

import (
	"errors"
	"reisen-be/internal/model"
	"reisen-be/internal/query"
	"reisen-be/internal/repository"
	"time"
)

type ContestService struct {
	contestListQuery *query.ContestListQuery
	contestRepo      *repository.ContestRepository
	problemRepo      *repository.ProblemRepository
	signupRepo       *repository.SignupRepository
	userRepo         *repository.UserRepository
	rankingRepo      *repository.RankingRepository
	JudgementRepo    *repository.JudgementRepository
}

func NewContestService(
	contestListQuery *query.ContestListQuery,
	contestRepo      *repository.ContestRepository,
	problemRepo      *repository.ProblemRepository,
	signupRepo       *repository.SignupRepository,
	userRepo         *repository.UserRepository,
	rankingRepo      *repository.RankingRepository,
	JudgementRepo    *repository.JudgementRepository,
) *ContestService {
	return &ContestService{
	  contestListQuery: contestListQuery,
		contestRepo:      contestRepo,
		problemRepo:      problemRepo,
		signupRepo:       signupRepo,
		userRepo:         userRepo,
		rankingRepo:      rankingRepo,
		JudgementRepo:    JudgementRepo,
	}
}

func (s *ContestService) CreateContest(contest *model.Contest) error {
	contest.CreatedAt = time.Now()
	contest.UpdatedAt = time.Now()
	return s.contestRepo.Create(contest)
}

func (s *ContestService) UpdateContest(contest *model.Contest) error {
	contest.UpdatedAt = time.Now()
	return s.contestRepo.Update(contest)
}

func (s *ContestService) GetContest(id model.ContestId) (*model.Contest, error) {
	return s.contestRepo.GetByID(id)
}

func (s *ContestService) GetSignup(id model.ContestId, userID model.UserId) (*model.Signup, error) {
	return s.signupRepo.GetSignup(userID, id)
}

func (s *ContestService) ListContests(filter *model.ContestFilter, userID *model.UserId, page, pageSize int) ([]model.ContestWithSignups, int64, error) {
	return s.contestListQuery.List(filter, userID, page, pageSize)
}

func (s *ContestService) AllContests(filter *model.ContestFilter, page, pageSize int) ([]model.Contest, int64, error) {
	return s.contestRepo.List(filter, page, pageSize)
}

func (s *ContestService) Signup(userID model.UserId, contestID model.ContestId) error {
	// 检查比赛是否已开始
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return err
	}

	if time.Now().After(contest.StartTime) {
		return errors.New("contest has already started")
	}

	return s.signupRepo.Signup(userID, contestID)
}

func (s *ContestService) Signout(userID model.UserId, contestID model.ContestId) error {
	// 检查比赛是否已开始
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return err
	}

	if time.Now().After(contest.StartTime) {
		return errors.New("contest has already started")
	}

	return s.signupRepo.Signout(userID, contestID)
}

func (s *ContestService) GetContestProblems(contestID model.ContestId) ([]model.ProblemCore, error) {
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return nil, err
	}

	var problemIDs []model.ProblemId
	for _, id := range contest.Problems {
		problemIDs = append(problemIDs, id)
	}

	problems := make([]model.ProblemCore, 0, len(problemIDs))
	for _, id := range problemIDs {
		problem, err := s.problemRepo.GetByID(id)
		if err != nil {
			continue // 跳过无效的题目
		}

		problems = append(problems, problem.ProblemCore)
	}

	return problems, nil
}

func (s *ContestService) GetRanking(contestID model.ContestId, userID model.UserId) (*model.Ranking, error) {
	return s.rankingRepo.GetByID(contestID, userID)
}

func (s *ContestService) GetRanklist(contestID model.ContestId) ([]model.Ranking, error) {
	return s.rankingRepo.GetByContest(contestID)
}

// 获取用户练习列表
func (s *ContestService) ListPractice(user model.UserId) ([]model.Ranking, error) {
	rankings, err := s.rankingRepo.GetByUser(user)
	if err != nil {
		return nil, err
	}
	return rankings, nil
}

// func (s *ContestService) UpdateJudgement(contestID model.ContestId, userID model.UserId, problemID model.ProblemId, Judgement model.Judgement) error {
// 	// 1. 更新题目结果
// 	err := s.JudgementRepo.UpdateJudgement(contestID, userID, problemID, Judgement)
// 	if err != nil {
// 		return err
// 	}

// 	// 2. 重新计算排名
// 	return s.calculateRankings(contestID)
// }

// func (s *ContestService) calculateRankings(contestID model.ContestId) error {
// 	// 获取比赛规则
// 	contest, err := s.contestRepo.GetByID(contestID)
// 	if err != nil {
// 		return err
// 	}

// 	// 获取所有报名用户
// 	signups, err := s.signupRepo.GetSignupsAll(contestID)
// 	if err != nil {
// 		return err
// 	}

// 	// 为每个用户计算排名数据
// 	var rankings []model.Ranking
// 	for _, reg := range signups {
// 		Judgements, err := s.JudgementRepo.GetJudgements(contestID, reg.UserID)
// 		if err != nil {
// 			return err
// 		}

// 		var totalScore int
// 		var totalPenalty int
// 		var solvedCount int

// 		for _, res := range Judgements {
// 			switch contest.Rule {
// 			case model.ContestRuleACM:
// 				if res.Judge == "correct" {
// 					solvedCount++
// 					totalPenalty += res.Penalty
// 				}
// 			case model.ContestRuleOI, model.ContestRuleIOI:
// 				if score, err := strconv.Atoi(res.Judge); err == nil {
// 					totalScore += score
// 				}
// 			}
// 		}

// 		rankings = append(rankings, model.Ranking{
// 			ContestID:  contestID,
// 			UserID:     reg.UserID,
// 			Judgements: Judgements,
// 			Ranking:       0, // 临时值，后面会重新计算
// 		})
// 	}

// 	// 根据比赛规则排序
// 	switch contest.Rule {
// 	case model.ContestRuleACM:
// 		// ACM规则：按解题数降序，罚时升序
// 		sort.Slice(rankings, func(i, j int) bool {
// 			if len(rankings[i].Judgements) != len(rankings[j].Judgements) {
// 				return len(rankings[i].Judgements) > len(rankings[j].Judgements)
// 			}
// 			return rankings[i].Judgements[0].Penalty < rankings[j].Judgements[0].Penalty
// 		})
// 	case model.ContestRuleOI, model.ContestRuleIOI:
// 		// OI / IOI 规则：按总分降序
// 		sort.Slice(rankings, func(i, j int) bool {
// 			var scoreI, scoreJ int
// 			for _, res := range rankings[i].Judgements {
// 				if s, err := strconv.Atoi(res.Judge); err == nil {
// 					scoreI += s
// 				}
// 			}
// 			for _, res := range rankings[j].Judgements {
// 				if s, err := strconv.Atoi(res.Judge); err == nil {
// 					scoreJ += s
// 				}
// 			}
// 			return scoreI > scoreJ
// 		})
// 	}

// 	// 分配最终排名（考虑并列情况）
// 	for i := range rankings {
// 		if i == 0 {
// 			rankings[i].Ranking = 1
// 		} else {
// 			if s.isRankEqual(contest.Rule, rankings[i-1], rankings[i]) {
// 				rankings[i].Ranking = rankings[i-1].Ranking
// 			} else {
// 				rankings[i].Ranking = i + 1
// 			}
// 		}
// 	}

// 	// 保存排名
// 	for _, ranking := range rankings {
// 		if err := s.rankingRepo.UpdateRanking(&ranking); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (s *ContestService) isRankEqual(rule model.ContestRule, a, b model.Ranking) bool {
// 	switch rule {
// 	case model.ContestRuleACM:
// 		return len(a.Judgements) == len(b.Judgements) && a.Judgements[0].Penalty == b.Judgements[0].Penalty
// 	case model.ContestRuleOI, model.ContestRuleIOI:
// 		var scoreA, scoreB int
// 		for _, res := range a.Judgements {
// 			if s, err := strconv.Atoi(res.Judge); err == nil {
// 				scoreA += s
// 			}
// 		}
// 		for _, res := range b.Judgements {
// 			if s, err := strconv.Atoi(res.Judge); err == nil {
// 				scoreB += s
// 			}
// 		}
// 		return scoreA == scoreB
// 	}
// 	return false
// }
