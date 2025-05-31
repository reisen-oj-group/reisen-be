package controller

import (
	"net/http"
	"reisen-be/internal/model"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
)

type ProblemController struct {
	problemService *service.ProblemService
}

func NewProblemController(problemService *service.ProblemService) *ProblemController {
	return &ProblemController{problemService: problemService}
}

// 创建或更新题目
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
	
	user := ctx.MustGet("user").(*model.User)

	var err error
	if req.Problem.ID == 0 {
		// 创建题目
		req.Problem.Provider = user.ID
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

// 获取题目详情
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

// 获取题目列表
func (c *ProblemController) ListProblems(ctx *gin.Context, isPublic bool, isMine bool) {
	var req model.ProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if isPublic {
		// 强制设置状态为 public
		req.Status = new(model.ProblemStatus)
		*req.Status = "public"
	} else {
		user := ctx.MustGet("user").(*model.User)

		if isMine {
			// 强制设置提供者为 user.ID
			req.Provider = new(model.UserId)
			*req.Provider = user.ID
		} else {
			if user.Role < model.RoleAdmin {
				ctx.JSON(http.StatusForbidden, gin.H{"error": "No permission"})
				return
			}
		}
	}

	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	pageSize := 50

	// 转换前端请求为过滤条件
	filter := req.ProblemFilter

	problems, total, err := c.problemService.ListProblems(&filter, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ProblemListResponse{
		Total:   int(total),
		Problems: problems,
	})
}

func (c *ProblemController) ListProblemsPublic(ctx *gin.Context) {
	c.ListProblems(ctx, true, false)
}

func (c *ProblemController) ListProblemsMine(ctx *gin.Context) {
	c.ListProblems(ctx, false, true)
}

func (c *ProblemController) ListProblemsAdmin(ctx *gin.Context) {
	c.ListProblems(ctx, false, false)
}

// 删除试题
func (c *ProblemController) DeleteProblem(ctx *gin.Context) {
	var req model.ProblemDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.problemService.DeleteProblem(req.Problem); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "problem deleted successfully"})
}
