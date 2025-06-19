package service

import (
	"encoding/json"
	"errors"
	"log"
	"reisen-be/internal/model"
	"reisen-be/internal/query"
	"reisen-be/internal/repository"
	"sort"
	"time"

	"gorm.io/datatypes"
)

type ContestService struct {
	contestListQuery *query.ContestListQuery
	contestRepo      *repository.ContestRepository
	problemRepo      *repository.ProblemRepository
	submissionRepo   *repository.SubmissionRepository
	signupRepo       *repository.SignupRepository
	userRepo         *repository.UserRepository
	rankingRepo      *repository.RankingRepository
	ticker           *time.Ticker
	stopChan         chan struct{}
}

func NewContestService(
	contestListQuery *query.ContestListQuery,
	contestRepo *repository.ContestRepository,
	problemRepo *repository.ProblemRepository,
	submissionRepo *repository.SubmissionRepository,
	signupRepo *repository.SignupRepository,
	userRepo *repository.UserRepository,
	rankingRepo *repository.RankingRepository,
	interval time.Duration,
) *ContestService {

	service := &ContestService{
		contestListQuery: contestListQuery,
		contestRepo:      contestRepo,
		problemRepo:      problemRepo,
		submissionRepo:   submissionRepo,
		signupRepo:       signupRepo,
		userRepo:         userRepo,
		rankingRepo:      rankingRepo,
	}

	service.StartRankingUpdater(interval)

	return service
}

func (s *ContestService) StartRankingUpdater(interval time.Duration) {
	s.ticker = time.NewTicker(interval)
	s.stopChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.ticker.C:
				log.Printf("Try to update ranking for contests")
				s.updateRunningContestsRankings()
			case <-s.stopChan:
				s.ticker.Stop()
				return
			}
		}
	}()
}

// 停止定时任务
func (s *ContestService) StopRankingUpdater() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
}

// 更新所有进行中比赛的榜单
func (s *ContestService) updateRunningContestsRankings() error {
	contests, err := s.contestRepo.ListRunning()
	if err != nil {
		return err
	}
	for _, contest := range contests {
		log.Printf("Try to update ranking for contest %d", contest.ID)

		if err := s.UpdateRankings(contest.ID); err != nil {
			log.Printf("Failed to update ranking for contest %d: %v", contest.ID, err)
			continue
		}
	}
	return nil
}

// 比赛结束时最终更新
func (s *ContestService) FinalizeContestRanking(contestID model.ContestId) error {
	// 等待所有提交完成
	for {
		hasPending, err := s.submissionRepo.CheckHasPending(contestID)
		if err != nil {
			return err
		}
		if !hasPending {
			break
		}
		time.Sleep(5 * time.Second) // 每 5 秒检查一次
	}
	// 最终更新榜单
	return s.UpdateRankings(contestID)
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

func (s *ContestService) DeleteContest(id model.ContestId) error {
	return s.contestRepo.Delete(id)
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

// 根据用户提交更新其 Ranking 的通过信息（不包括排名）
func (s *ContestService) UpdateRanking(submission *model.Submission) error {
	// 只处理比赛提交
	if submission.ContestID == nil {
		return nil
	}

	contestID := *submission.ContestID

	// 获取比赛信息
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return err
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(submission.UserID)
	if err != nil {
		return err
	}

	// 获取原先 Ranking 值
	ranking, err := s.rankingRepo.GetByID(contestID, submission.UserID)
	if err != nil {
		// 若不存在则新建
		ranking = &model.Ranking{
			ContestID: contestID,
			UserID:    submission.UserID,
			Team:      user.Name,
		}

		// 根据赛制进行初始化
		switch contest.Rule {
		case model.ContestRuleACM:
			ranking.Detail, _ = json.Marshal(model.ACMDetail{
				Type:         "ACM",
				TotalPenalty: 0,
				TotalSolved:  0,
				Problems:     make(map[model.ProblemId]model.ACMCell),
			})
		case model.ContestRuleOI:
			ranking.Detail, _ = json.Marshal(model.OIDetail{
				Type:       "OI",
				TotalScore: 0,
				Problems:   make(map[model.ProblemId]model.OIProblem),
			})
		case model.ContestRuleIOI:
			ranking.Detail, _ = json.Marshal(model.IOIDetail{
				Type:       "IOI",
				TotalScore: 0,
				Problems:   make(map[model.ProblemId]model.IOIProblem),
			})
		}
	}

	// 查询是否是比赛内的试题

	var exist bool = false
	for _, id := range contest.Problems {
		if id == submission.ProblemID {
			exist = true
			break
		}
	}
	if !exist {
		return errors.New("problem not found in contest")
	}

	// 根据比赛规则更新信息
	switch contest.Rule {
	case model.ContestRuleACM:
		return s.updateACMRanking(contest, ranking, submission)
	case model.ContestRuleOI:
		return s.updateOIRanking(contest, ranking, submission)
	case model.ContestRuleIOI:
		return s.updateIOIRanking(contest, ranking, submission)
	}

	return nil
}

// 处理 ACM 赛制的排名信息
func (s *ContestService) updateACMRanking(contest *model.Contest, ranking *model.Ranking, submission *model.Submission) error {

	var detail model.ACMDetail
	if err := json.Unmarshal(ranking.Detail, &detail); err != nil {
		return err
	}

	problemID := submission.ProblemID
	problem := detail.Problems[problemID]

	// Only process if not already solved (ACM rules)
	if !problem.IsSolved {
		// Check if submission was before freeze time
		isBeforeFreeze := contest.EndTime.Sub(submission.SubmittedAt) > time.Hour // Assuming 1 hour freeze

		if isBeforeFreeze {
			problem.AttemptBF++
		} else {
			problem.AttemptAF++
		}

		if submission.Verdict == model.VerdictAC {
			problem.IsSolved = true
			problem.Penalty = (problem.AttemptBF + problem.AttemptAF - 1) * 20

			// Calculate time in minutes from contest start
			problem.Penalty += int(submission.SubmittedAt.Sub(contest.StartTime).Minutes())

			// Check if this is first AC (blood)
			if contest.ProblemStatus == nil {
				contest.ProblemStatus = make(map[model.ProblemId]model.ContestProblemStatus)
			}

			if status, ok := contest.ProblemStatus[problemID]; !ok || status.FirstBloodUserID == nil {
				problem.IsFirst = true

				// Update contest problem status
				contest.ProblemStatus[problemID] = model.ContestProblemStatus{
					FirstBloodUserID: &ranking.UserID,
					FirstBloodTime:   &submission.SubmittedAt,
					SolvedCount:      1,
				}

				// Save contest update
				if err := s.contestRepo.Update(contest); err != nil {
					return err
				}
			} else {
				// Increment solved count
				status.SolvedCount++
				contest.ProblemStatus[problemID] = status
				if err := s.contestRepo.Update(contest); err != nil {
					return err
				}
			}

			// Update totals
			detail.TotalSolved++
			detail.TotalPenalty += problem.Penalty
		}

		// Update the problem data
		detail.Problems[problemID] = problem

		// Marshal back to JSON
		updatedDetail, err := json.Marshal(detail)
		if err != nil {
			return err
		}
		ranking.Detail = datatypes.JSON(updatedDetail)

		// Save the ranking
		if ranking.Ranking == 0 {
			return s.rankingRepo.Create(ranking)
		}
		return s.rankingRepo.Update(ranking)
	}

	return nil
}

// updateOIRanking handles OI contest ranking updates
func (s *ContestService) updateOIRanking(_ *model.Contest, ranking *model.Ranking, submission *model.Submission) error {

	var detail model.OIDetail
	if err := json.Unmarshal(ranking.Detail, &detail); err != nil {
		return err
	}

	problemID := submission.ProblemID

	// Initialize problem if not exists
	if _, ok := detail.Problems[problemID]; !ok {
		detail.Problems[problemID] = model.OIProblem{
			Score: 0,
		}
	}

	// Always update score in OI mode
	problem := detail.Problems[problemID]
	if submission.Score != nil {
		problem.Score = *submission.Score
	}

	detail.Problems[problemID] = problem

	// Update totals
	total := 0
	for _, p := range detail.Problems {
		total += p.Score
	}
	detail.TotalScore = total

	// Marshal back to JSON
	updatedDetail, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	ranking.Detail = datatypes.JSON(updatedDetail)

	// Save the ranking
	if ranking.Ranking == 0 {
		return s.rankingRepo.Create(ranking)
	}
	return s.rankingRepo.Update(ranking)
}

// updateIOIRanking handles IOI contest ranking updates (similar to OI for now)
func (s *ContestService) updateIOIRanking(_ *model.Contest, ranking *model.Ranking, submission *model.Submission) error {

	var detail model.IOIDetail
	if err := json.Unmarshal(ranking.Detail, &detail); err != nil {
		return err
	}
	problemID := submission.ProblemID

	// Initialize problem if not exists
	if _, ok := detail.Problems[problemID]; !ok {
		detail.Problems[problemID] = model.IOIProblem{
			Score: 0,
		}
	}

	// Always update score in IOI mode
	problem := detail.Problems[problemID]
	if submission.Score != nil {
		problem.Score = *submission.Score
	}

	// Update totals
	total := 0
	for _, p := range detail.Problems {
		total += p.Score
	}
	detail.TotalScore = total

	// Marshal back to JSON
	updatedDetail, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	ranking.Detail = datatypes.JSON(updatedDetail)

	// Save the ranking
	if ranking.Ranking == 0 {
		return s.rankingRepo.Create(ranking)
	}
	return s.rankingRepo.Update(ranking)
}

// UpdateRankings recalculates rankings for all participants in a contest
func (s *ContestService) UpdateRankings(contestID model.ContestId) error {
	// Get contest info
	contest, err := s.contestRepo.GetByID(contestID)
	if err != nil {
		return err
	}

	// Get all rankings for the contest
	rankings, err := s.rankingRepo.GetByContest(contestID)
	if err != nil {
		return err
	}

	// Prepare ranking data for sorting
	type rankData struct {
		index     int
		ranking   *model.Ranking
		sortValue interface{}
	}

	rankInfos := make([]rankData, len(rankings))

	// Calculate sorting values based on contest type
	for i, ranking := range rankings {
		rankInfos[i] = rankData{
			index:   i,
			ranking: &ranking,
		}

		switch contest.Rule {
		case model.ContestRuleACM:
			var detail model.ACMDetail
			if err := json.Unmarshal(ranking.Detail, &detail); err != nil {
				continue
			}
			rankInfos[i].sortValue = struct {
				solved  int
				penalty int
			}{
				solved:  detail.TotalSolved,
				penalty: detail.TotalPenalty,
			}

		case model.ContestRuleOI, model.ContestRuleIOI:
			var detail model.OIDetail // Can use OIDetail for both since structure is same
			if err := json.Unmarshal(ranking.Detail, &detail); err != nil {
				continue
			}
			rankInfos[i].sortValue = detail.TotalScore
		}
	}

	// Sort based on contest type
	switch contest.Rule {
	case model.ContestRuleACM:
		sort.Slice(rankInfos, func(i, j int) bool {
			iVal := rankInfos[i].sortValue.(struct {
				solved  int
				penalty int
			})
			jVal := rankInfos[j].sortValue.(struct {
				solved  int
				penalty int
			})

			if iVal.solved != jVal.solved {
				return iVal.solved > jVal.solved // More solved first
			}
			return iVal.penalty < jVal.penalty // Lower penalty first
		})

	case model.ContestRuleOI, model.ContestRuleIOI:
		sort.Slice(rankInfos, func(i, j int) bool {
			return rankInfos[i].sortValue.(int) > rankInfos[j].sortValue.(int) // Higher score first
		})
	}

	// Update rankings with new positions
	currentRank := 1
	for i, rankInfo := range rankInfos {
		// Handle ties
		if i > 0 {
			prev := rankInfos[i-1]
			switch contest.Rule {
			case model.ContestRuleACM:
				iVal := rankInfo.sortValue.(struct {
					solved  int
					penalty int
				})
				prevVal := prev.sortValue.(struct {
					solved  int
					penalty int
				})

				if iVal.solved == prevVal.solved && iVal.penalty == prevVal.penalty {
					rankInfo.ranking.Ranking = prev.ranking.Ranking
					currentRank++
					continue
				}

			case model.ContestRuleOI, model.ContestRuleIOI:
				if rankInfo.sortValue.(int) == prev.sortValue.(int) {
					rankInfo.ranking.Ranking = prev.ranking.Ranking
					currentRank++
					continue
				}
			}
		}

		rankInfo.ranking.Ranking = currentRank
		currentRank++
	}

	// Save all updated rankings
	for _, rankInfo := range rankInfos {
		if err := s.rankingRepo.Update(rankInfo.ranking); err != nil {
			return err
		}
	}

	return nil
}
