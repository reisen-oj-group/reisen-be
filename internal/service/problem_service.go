package service

import (
	"reisen-be/internal/filesystem"
	"reisen-be/internal/model"
	"reisen-be/internal/query"
	"reisen-be/internal/repository"
)

type ProblemService struct {
	problemListQuery  *query.ProblemListQuery
	problemRepo       *repository.ProblemRepository
	problemFilesystem *filesystem.ProblemFilesystem
}

func NewProblemService(
	problemListQuery *query.ProblemListQuery,
	problemRepo *repository.ProblemRepository,
	problemFilesystem *filesystem.ProblemFilesystem,
) *ProblemService {
	return &ProblemService{
		problemListQuery:  problemListQuery,
		problemRepo:       problemRepo,
		problemFilesystem: problemFilesystem,
	}
}

func (s *ProblemService) CreateProblem(problem *model.Problem) error {
	return s.problemRepo.Create(problem)
}

func (s *ProblemService) UpdateProblem(problem *model.Problem) error {
	return s.problemRepo.Update(problem)
}

func (s *ProblemService) GetProblem(id model.ProblemId) (*model.Problem, error) {
	return s.problemRepo.GetByID(id)
}

func (s *ProblemService) AllProblems(filter *model.ProblemFilter, page, pageSize int) ([]model.ProblemCore, int64, error) {
	return s.problemRepo.List(filter, page, pageSize)
}

func (s *ProblemService) ListProblems(filter *model.ProblemFilter, userID *model.UserId, page, pageSize int) ([]model.ProblemCoreWithJudgements, int64, error) {
	return s.problemListQuery.List(filter, userID, page, pageSize)
}

func (s *ProblemService) DeleteProblem(id model.ProblemId) error {
	return s.problemRepo.Delete(id)
}

func (s *ProblemService) UploadTestdata(problemID model.ProblemId, filePath string) error {
	problem, err := s.problemRepo.GetByID(problemID)
	if err != nil {
		return err
	}
	// 上传测试数据
	if err := s.problemFilesystem.UploadTestdata(problemID, filePath); err != nil {
		return err
	}
	// 生成配置文件
	if err := s.problemFilesystem.GenerateConfig(problem.ProblemCore); err != nil {
		return err
	}

	// 更新数据库记录
	return s.problemRepo.UpdateTestdataStatus(problemID, true, false)
}

func (s *ProblemService) DownloadTestdata(problemID model.ProblemId) (*string, error) {
	return s.problemFilesystem.DownloadTestdata(problemID)
}

func (s *ProblemService) DeleteTestdata(problemID model.ProblemId) error {
	if err := s.problemFilesystem.DeleteTestdata(problemID); err != nil {
		return err
	}
	return s.problemRepo.UpdateTestdataStatus(problemID, false, false)
}

func (s *ProblemService) UploadConfig(problemID model.ProblemId, config *model.TestdataConfig) error {
	err := s.problemFilesystem.UploadConfig(problemID, config)
	if err != nil {
		return err
	}
	// 更新数据库记录
	return s.problemRepo.UpdateTestdataStatus(problemID, true, true)
}

func (s *ProblemService) GetConfig(problemID model.ProblemId) (*model.TestdataConfig, error) {
	return s.problemFilesystem.GetConfig(problemID)
}
