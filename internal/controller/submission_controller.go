package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

type SubmissionController struct {
	judgeService *service.JudgeService
}

func NewSubmissionController(judgeService *service.JudgeService) *SubmissionController {
	return &SubmissionController{judgeService: judgeService}
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

// 获取评测记录列表
func (c *SubmissionController) ListSubmissions(ctx *gin.Context) {
	var req model.SubmissionListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pageSize := 50
	submissions, total, err := c.judgeService.ListSubmissions(&req.SubmissionFilterParamsRaw, req.Page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.SubmissionListResponse{
		Total:   int(total),
		Submissions: submissions,
	})
}