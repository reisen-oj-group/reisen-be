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

// 提交代码评测
func (c *RecordController) SubmitCode(ctx *gin.Context) {
	var req model.JudgeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取用户
	user := ctx.MustGet("user").(*model.User)

	record, err := c.judgeService.SubmitCode(&req, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.JudgeResponse{
		Record: record.ID,
	})
}

// 获取评测记录详情
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

// 获取评测记录列表
func (c *RecordController) ListRecords(ctx *gin.Context) {
	var req model.RecordListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pageSize := 50
	records, total, err := c.judgeService.ListSubmissions(&req.RecordFilterParamsRaw, req.Page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.RecordListResponse{
		Total:   int(total),
		Records: records,
	})
}