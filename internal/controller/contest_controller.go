package controller

import (
	"errors"
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContestController struct {
	contestService *service.ContestService
	problemService *service.ProblemService
	userService    *service.UserService
	judgeService   *service.JudgeService
}

func NewContestController(
	contestService *service.ContestService,
	problemService *service.ProblemService,
	userService    *service.UserService,
	judgeService   *service.JudgeService,
) *ContestController {
	return &ContestController{
		contestService: contestService,
		problemService: problemService,
		userService:    userService,
		judgeService:   judgeService,
	}
}

// 创建或更新比赛
func (c *ContestController) CreateOrUpdateContest(ctx *gin.Context) {
	var req model.ContestEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	if req.Contest.ID == 0 {
		err = c.contestService.CreateContest(&req.Contest)
	} else {
		err = c.contestService.UpdateContest(&req.Contest)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestEditResponse{
		Contest: req.Contest,
	})
}

// 获取比赛详情
func (c *ContestController) GetContest(ctx *gin.Context) {
	var req model.ContestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contest, err := c.contestService.GetContest(req.Contest)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "比赛不存在"})
		return
	}

	user := ctx.MustGet("user").(*model.User)
	
	if contest.StartTime.After(time.Now()) && (user == nil || user.Role <= model.RoleUser) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "比赛未开始"})
		return
	}

	if user == nil {
		// 游客，只有比赛信息
		ctx.JSON(http.StatusOK, model.ContestResponse{
			Contest: *contest,
		})
	} else {
		signup, _ := c.contestService.GetSignup(req.Contest, user.ID)
		ranking, _ := c.contestService.GetRanking(req.Contest, user.ID)
		ctx.JSON(http.StatusOK, model.ContestResponse{
			Contest: *contest,
			Signup:  signup,
			Ranking: ranking,
		})
	}
}

// 删除比赛
func (c *ContestController) DeleteContest(ctx *gin.Context) {
	var req model.ContestDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.contestService.DeleteContest(req.Contest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "contest deleted successfully"})
}

// 提交主题库代码评测
func (c *ContestController) SubmitCode(ctx *gin.Context) {
	var req model.JudgeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取用户
	user := ctx.MustGet("user").(*model.User)

	// 获取比赛信息
	contest, err := c.contestService.GetContest(*req.Contest)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if contest.StartTime.After(time.Now()) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "比赛未开始"})
		return
	}
	if contest.EndTime.Before(time.Now()) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "比赛已结束"})
		return
	}

	// 获取报名信息
	_, err = c.contestService.GetSignup(*req.Contest, user.ID)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "报名后才可提交比赛"})
		return
	}

	submission, err := c.judgeService.SubmitCode(&req, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.JudgeResponse{
		Submission: submission.ID,
	})
}

// 获取比赛排名
func (c *ContestController) GetRanking(ctx *gin.Context) {
	var req model.ContestRankingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := ctx.MustGet("user").(*model.User)

	ranking, err := c.contestService.GetRanking(req.Contest, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound){
			ctx.JSON(http.StatusOK, model.ContestRankingResponse{
				Ranking: nil,
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, model.ContestRankingResponse{
		Ranking: ranking,
	})
}

// 将原始参数转换为处理后的参数
func (c *ContestController) ConvertContestFilter(raw *model.ContestFilterRaw) (*model.ContestFilter, error) {
	if raw == nil {
		return nil, nil
	}
	params := &model.ContestFilter{
		Status:     raw.Status,
		Rule:       raw.Rule,
		Difficulty: raw.Difficulty,
		Keyword:    raw.Keyword,
		Before:     raw.Before,
		After:      raw.After,
	}

	return params, nil
}

// 获取比赛列表
func (c *ContestController) ListContests(ctx *gin.Context) {
	var req model.ContestListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := ctx.MustGet("user").(*model.User)
	var userID *model.UserId

	if user == nil {
		userID = nil
	} else {
		userID = &user.ID
	}

	filter, err := c.ConvertContestFilter(&req.ContestFilterRaw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	size := 10

	contests, total, err := c.contestService.ListContests(filter, userID, page, size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestListResponse{
		Total:    total,
		Contests: contests,
	})
}

// 获取比赛列表
func (c *ContestController) AllContests(ctx *gin.Context) {
	var req model.ContestAllRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter, err := c.ConvertContestFilter(&req.ContestFilterRaw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	size := 50
	if req.Size != nil && *req.Size > 0 {
		size = *req.Size
	}

	contests, total, err := c.contestService.AllContests(filter, page, size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestAllResponse{
		Total:    total,
		Contests: contests,
	})
}

// 获取注册情况
func (c *ContestController) GetSignup(ctx *gin.Context) {
	var req model.ContestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := ctx.MustGet("user").(*model.User)
	resp, err := c.contestService.GetSignup(req.Contest, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// 报名比赛
func (c *ContestController) SignupContest(ctx *gin.Context) {
	var req model.ContestSignupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := ctx.MustGet("user").(*model.User)
	if err := c.contestService.Signup(user.ID, req.Contest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestSignupResponse{})
}

// 取消报名比赛
func (c *ContestController) SignoutContest(ctx *gin.Context) {
	var req model.ContestSignoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := ctx.MustGet("user").(*model.User)
	if err := c.contestService.Signout(user.ID, req.Contest); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestSignoutResponse{})
}

// 获取比赛题目集
func (c *ContestController) GetContestProblems(ctx *gin.Context) {
	var req model.ContestProblemsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	problems, err := c.contestService.GetContestProblems(req.Contest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestProblemsResponse{
		Problems: problems,
	})
}

// 获取比赛排行榜
func (c *ContestController) GetRanklist(ctx *gin.Context) {
	var req model.ContestRanklistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rankings, err := c.contestService.GetRanklist(req.Contest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ContestRanklistResponse{
		Rankings: rankings,
	})
}

// // 更新题目结果（供评测系统调用）
// func (c *ContestController) UpdateResult(ctx *gin.Context) {
// 	var req model.ResultUpdateRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if err := c.contestService.UpdateResult(req.Contest, req.User, req.Problem, req.Result); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ranking, err := c.contestService.GetRanking(req.Contest, req.User)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, model.ResultUpdateResponse{
// 		Ranking: *ranking,
// 	})
// }
