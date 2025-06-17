package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

type SubmissionController struct {
	judgeService *service.JudgeService
	userService *service.UserService
}

func NewSubmissionController(
	judgeService *service.JudgeService,
	userService *service.UserService,
) *SubmissionController {
	return &SubmissionController{
		judgeService: judgeService,
		userService: userService,
	}
}

// 提交代码评测
func (c *SubmissionController) SubmitCode(ctx *gin.Context) {
	var req model.JudgeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取用户
	user := ctx.MustGet("user").(*model.User)

	submission, err := c.judgeService.SubmitCode(&req, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.JudgeResponse{
		Submission: submission.ID,
	})
}

// 获取评测记录详情
func (c *SubmissionController) GetSubmissionDetail(ctx *gin.Context) {
	var req model.SubmissionDetailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submission, err := c.judgeService.GetSubmissionDetail(req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.SubmissionDetailResponse{
		Submission: *submission,
	})
}



// 将原始参数转换为处理后的参数
func (c *SubmissionController) ConvertSubmissionFilter(raw *model.SubmissionFilterRaw) (*model.SubmissionFilter, error) {
	if raw == nil {
		return nil, nil
	}

	params := &model.SubmissionFilter{
			Problem: raw.Problem,
			Lang:    raw.Lang,
			Verdict: raw.Verdict,
	}
	// 处理 User 字段转换
	if raw.User != nil {
		userID, err := c.userService.ParseUsername(*raw.User)
		if err != nil {
				return nil, err
		}
		params.User = userID
	}
	return params, nil
}

// 获取评测记录列表
func (c *SubmissionController) ListSubmissions(ctx *gin.Context) {
	var req model.SubmissionListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter, err := c.ConvertSubmissionFilter(&req.SubmissionFilterRaw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pageSize := 50
	submissions, total, err := c.judgeService.ListSubmissions(filter, req.Page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.SubmissionListResponse{
		Total:   total,
		Submissions: submissions,
	})
}

// 获取评测记录列表
func (c *SubmissionController) AllSubmissions(ctx *gin.Context) {
	c.ListSubmissions(ctx)
}