package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

type RecordController struct {
	judgeService *service.JudgeService
}

func NewRecordController(judgeService *service.JudgeService) *RecordController {
	return &RecordController{judgeService: judgeService}
}

// SubmitCode 提交代码评测
// @Summary 提交代码评测
// @Description 提交代码进行评测
// @Tags 评测记录
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body model.JudgeRequest true "评测请求"
// @Success 200 {object} model.JudgeResponse
// @Router /api/problem/submit [post]
func (c *RecordController) SubmitCode(ctx *gin.Context) {
	var req model.JudgeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取用户
	user := ctx.MustGet("user").(model.User)

	record, err := c.judgeService.SubmitCode(&req, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.JudgeResponse{
		Record: record.ID,
	})
}

// GetRecordDetail 获取评测记录详情
// @Summary 获取评测记录详情
// @Description 根据ID获取评测记录详情
// @Tags 评测记录
// @Accept json
// @Produce json
// @Param request body model.RecordDetailRequest true "记录ID"
// @Success 200 {object} model.RecordDetailResponse
// @Router /api/record/detail [post]
func (c *RecordController) GetRecordDetail(ctx *gin.Context) {
	var req model.RecordDetailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := c.judgeService.GetSubmissionDetail(req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.RecordDetailResponse{
		Record: *record,
	})
}

// ListRecords 获取评测记录列表
// @Summary 获取评测记录列表
// @Description 获取评测记录列表，支持分页和过滤
// @Tags 评测记录
// @Accept json
// @Produce json
// @Param request body model.RecordListRequest true "过滤条件"
// @Success 200 {object} model.RecordListResponse
// @Router /api/record/list [post]
func (c *RecordController) ListRecords(ctx *gin.Context) {
	var req model.RecordListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pageSize := 50
	records, total, err := c.judgeService.ListSubmissions(&req.RecordFilterParams, req.Page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.RecordListResponse{
		Total:   int(total),
		Records: records,
	})
}