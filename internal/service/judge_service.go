package service

import (
	"context"
	"fmt"
	"reisen-be/internal/filesystem"
	"reisen-be/internal/model"
	"reisen-be/internal/repository"
	"reisen-be/internal/service/judge"
	"reisen-be/internal/websocket"
	"strconv"
	"time"
)

type JudgeService struct {
	submissionRepo     *repository.SubmissionRepository
	judgementRepo      *repository.JudgementRepository
	problemRepo        *repository.ProblemRepository
	userRepo           *repository.UserRepository
	
	dispatcher         *judge.Dispatcher
	problemFilesystem  *filesystem.ProblemFilesystem
	contestService     *ContestService
}

func NewJudgeService(
    submissionRepo *repository.SubmissionRepository,
    judgementRepo *repository.JudgementRepository,
    problemRepo *repository.ProblemRepository,
    userRepo *repository.UserRepository,
		problemFilesystem * filesystem.ProblemFilesystem,
	  submissionWs      *websocket.SubmissionWs,
		contestService     *ContestService,
    workers int,
) *JudgeService {
    compiler := judge.NewCompiler()
    runner := judge.NewRunner()
    
    // 默认使用严格判分器，实际会根据题目配置选择
    checker, _ := judge.NewChecker(model.JudgeConfig{CheckerType: "loose"})
    dispatcher := judge.NewDispatcher(workers, compiler, runner, checker, problemFilesystem, submissionWs)
    
    ctx := context.Background()
    dispatcher.Start(ctx)

    s := &JudgeService{
        submissionRepo:     submissionRepo,
        judgementRepo:      judgementRepo,
        problemRepo:        problemRepo,
        userRepo:           userRepo,
        dispatcher:         dispatcher,
        problemFilesystem:  problemFilesystem,
				contestService:     contestService,
    }
		return s
}


// 根据用户提交更新其 Judgement 的通过信息
func (s *JudgeService) UpdateJudgement(submission *model.Submission) error {

	// 获取用户信息
	problem, err := s.problemRepo.GetByID(submission.ProblemID)
	if err != nil {
		return err
	}

	// 获取原先 Ranking 值
	judgement, err := s.judgementRepo.GetByID(submission.ProblemID, submission.UserID)
	if err != nil {
		// 若不存在则新建
		judgement = &model.Judgement{
			ProblemID:  submission.ProblemID,
			UserID:     submission.UserID,
			Judge:      "incorrect",
			Difficulty: problem.Difficulty,
			Stamp:      nil,
		}
	}

	if judgement.Judge != "correct" {
			
		if submission.Verdict == model.VerdictAC {
			now := time.Now()
			judgement.Judge = "correct"
			judgement.Stamp = &now
		} else 
		if submission.Score != nil {
			flag := false
			if judgement.Judge != "incorrect" {
				curScore, err := strconv.Atoi(judgement.Judge)
				if err != nil {
					return err
				}
				if *submission.Score > curScore {
					flag = true
				}
			} else {
				flag = true
			}
			if flag {
				judgement.Judge = strconv.Itoa(*submission.Score)
			}
		}
	}

	s.judgementRepo.Update(judgement)
	return nil
}

func (s *JudgeService) processResult() error {
	// 等待结果
	submission := <-s.dispatcher.Results()

	// 保存更新
	if err := s.submissionRepo.Update(submission); err != nil {
		return err
	}

	s.UpdateJudgement(submission)
	s.contestService.UpdateRanking(submission)

	// 更新题目统计信息
	if submission.Verdict == model.VerdictAC {
		if err := s.problemRepo.IncreaseSubmitCorrect(submission.ProblemID); err != nil {
			return err
		}
	}
	if err := s.problemRepo.IncreaseSubmitTotal(submission.ProblemID); err != nil {
		return err
	}
	return nil
}


func (s *JudgeService) SubmitCode(req *model.JudgeRequest, userID model.UserId) (*model.SubmissionFull, error) {
	// 1. 获取题目信息
	problem, err := s.problemRepo.GetByID(req.Problem)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 3. 获取测试用例配置
	config, err := s.problemFilesystem.GetJudgeConfig(req.Problem)
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %v", err)
	}

	// 4. 创建初始提交记录
	now := time.Now()
	submission := model.Submission{
		SubmissionCore: model.SubmissionCore{
			ProblemID:   req.Problem,
			UserID:      userID,
			ContestID:   req.Contest,
			SubmittedAt: now,
			ProcessedAt: now,
			Lang:        req.Lang,
			CodeLength:  len(req.Code),
			Verdict:     model.VerdictPD, // Pending
		},
		Code:      req.Code,
		Testcases: make([]model.Testcase, len(config.TestCases)),
	}
	
	for i := range submission.Testcases {
		submission.Testcases[i].ID = i + 1
		submission.Testcases[i].Verdict = model.VerdictPD
	}

	// 5. 保存初始提交记录
	if err := s.submissionRepo.Create(&submission); err != nil {
		return nil, err
	}

	// 6. 准备评测任务
	task := &model.JudgeTask{
		Submission: submission,
		Config: model.JudgeConfig{
			TimeLimit:   problem.LimitTime,
			MemoryLimit: problem.LimitMemory,
			TestCases:   config.TestCases,
			// 以后可以从题目配置中读取判分器类型
			CheckerType: "loose",
		},
	}

	// 7. 提交评测任务
	s.dispatcher.Submit(task)

	// 8. 异步处理评测结果
	go s.processResult()

	// 9. 返回初始响应
	return &model.SubmissionFull{
		Submission: submission,
		Problem:    problem.ProblemCore,
		User:       *user,
	}, nil
}

// 获取提交详情
func (s *JudgeService) GetSubmissionDetail(id model.SubmissionId) (*model.SubmissionFull, error) {
	submission, err := s.submissionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	problem, err := s.problemRepo.GetByID(submission.ProblemID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(submission.UserID)
	if err != nil {
		return nil, err
	}

	return &model.SubmissionFull{
		Submission: *submission,
		Problem:    problem.ProblemCore,
		User:       *user,
	}, nil
}

// 获取提交列表
func (s *JudgeService) ListSubmissions(filter *model.SubmissionFilter, page, pageSize int) ([]model.SubmissionLite, int64, error) {
	submissions, total, err := s.submissionRepo.List(filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var lites []model.SubmissionLite
	for _, sub := range submissions {
		problem, err := s.problemRepo.GetByID(sub.ProblemID)
		if err != nil {
			return nil, 0, err
		}

		user, err := s.userRepo.GetByID(sub.UserID)
		if err != nil {
			return nil, 0, err
		}

		lites = append(lites, model.SubmissionLite{
			SubmissionCore: sub.SubmissionCore,
			Problem:        problem.ProblemCore,
			User:           *user,
		})
	}

	return lites, total, nil
}

// 获取用户练习列表
func (s *JudgeService) ListPractice(user model.UserId) ([]model.Judgement, error) {
	judgements, err := s.judgementRepo.GetByUser(user)
	if err != nil {
		return nil, err
	}
	return judgements, nil
}
