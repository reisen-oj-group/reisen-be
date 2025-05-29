package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProblemController struct {
	problemService *service.ProblemService
}

func NewProblemController(problemService *service.ProblemService) *ProblemController {
	return &ProblemController{problemService: problemService}
}

// CreateOrUpdateProblem 创建或更新题目
// @Summary 创建或更新题目
// @Description 创建或更新题目信息
// @Tags 题目管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body model.ProblemEditRequest true "题目信息"
// @Success 200 {object} model.ProblemEditResponse
// @Router /api/problem/edit [post]
func (c *ProblemController) CreateOrUpdateProblem(ctx *gin.Context) {
	var req model.ProblemEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换标签为关联模型
	var tags []model.ProblemTag
	for _, tag := range req.Problem.Tags {
		tags = append(tags, model.ProblemTag{TagID: tag.TagID})
	}
	req.Problem.Tags = tags

	var err error
	if req.Problem.ID == 0 {
		// 创建题目
		err = c.problemService.CreateProblem(&req.Problem)
	} else {
		// 更新题目
		err = c.problemService.UpdateProblem(&req.Problem)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ProblemEditResponse{
		Problem: req.Problem,
	})
}

// GetProblem 获取题目详情
// @Summary 获取题目详情
// @Description 根据ID获取题目详情
// @Tags 题目管理
// @Accept json
// @Produce json
// @Param request body model.ProblemRequest true "题目ID"
// @Success 200 {object} model.ProblemResponse
// @Router /api/problem [post]
func (c *ProblemController) GetProblem(ctx *gin.Context) {
    var req model.ProblemRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    problem, err := c.problemService.GetProblem(req.Problem)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 这里简化处理，实际应从数据库查询用户对该题目的结果
    var result *model.Result
    if req.User != nil {
        result = &model.Result{
            Problem: req.Problem,
            User:    *req.User,
            Judge:   "correct", // 示例值
            Time:    100,       // 示例值
        }
    }

    ctx.JSON(http.StatusOK, model.ProblemResponse{
        Problem: *problem,
        Result:  result,
    })
}

// ListProblems 获取题目列表
// @Summary 获取题目列表
// @Description 获取题目列表，支持分页和过滤
// @Tags 题目管理
// @Accept json
// @Produce json
// @Param request body model.ProblemListRequest true "过滤条件"
// @Success 200 {object} model.ProblemListResponse
// @Router /api/problem/list [post]
func (c *ProblemController) ListProblems(ctx *gin.Context) {
	var req model.ProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	pageSize := 50

	// 转换前端请求为过滤条件
	filter := &model.ProblemFilter{
		MinDifficulty: req.MinDifficulty,
		MaxDifficulty: req.MaxDifficulty,
		Tags:          req.Tags,
		Keywords:      req.Keywords,
	}

	problems, total, err := c.problemService.ListProblems(filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ProblemListResponse{
		Total:   int(total),
		Problems: problems,
	})
}

func (c *ProblemController) DeleteProblem(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	if err := c.problemService.DeleteProblem(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "problem deleted successfully"})
}
