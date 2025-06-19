package controller

import (
	"net/http"
	"os"
	"reisen-be/internal/model"
	"reisen-be/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProblemController struct {
	problemService *service.ProblemService
	judgeService *service.JudgeService
}

func NewProblemController(
	problemService *service.ProblemService, 
	judgeService *service.JudgeService,
) *ProblemController {
	return &ProblemController{
		problemService: problemService,
		judgeService: judgeService,
	}
}

// 创建或更新题目
func (c *ProblemController) CreateOrUpdateProblem(ctx *gin.Context) {
	var req model.ProblemEditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
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
	var Judgement *model.Judgement
	// if req.User != nil {
	// 	Judgement = &model.Judgement{
	// 		Problem: req.Problem,
	// 		User:    *req.User,
	// 		Judge:   "correct", // 示例值
	// 		Stamp:   time.now(), // 示例值
	// 	}
	// }

	ctx.JSON(http.StatusOK, model.ProblemResponse{
		Problem: *problem,
		Judgement:  Judgement,
	})
}

// 获取后台或者私人题目列表
func (c *ProblemController) allProblems(ctx *gin.Context, isMine bool) {
	var req model.ProblemAllRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	size := 50
	if req.Size != nil && *req.Size > 0 {
		size = *req.Size
	}

	// 转换前端请求为过滤条件
	filter := req.ProblemFilter

	problems, total, err := c.problemService.AllProblems(&filter, page, size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ProblemAllResponse{
		Total:    total,
		Problems: problems,
	})
}

// 获取主题库题目列表
func (c *ProblemController) listProblems(ctx *gin.Context) {
	var req model.ProblemListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	size := 50

	// 转换前端请求为过滤条件
	filter := req.ProblemFilter

	user := ctx.MustGet("user").(*model.User)
	var userID *model.UserId

	if user == nil {
		userID = nil
	} else {
		userID = &user.ID
	}

	problems, total, err := c.problemService.ListProblems(&filter, userID, page, size)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ProblemListResponse{
		Total:    total,
		Problems: problems,
	})
}

// 提交主题库代码评测
func (c *ProblemController) SubmitCode(ctx *gin.Context) {
	var req model.JudgeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 获取题目信息
	problem, err := c.problemService.GetProblem(req.Problem)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文中获取用户
	user := ctx.MustGet("user").(*model.User)

	// 普通用户只能提交公开试题
	if user.Role == model.RoleUser {
		if problem.Status != model.ProblemStatusPublic {
			ctx.Status(http.StatusForbidden)
			return
		}
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

func (c *ProblemController) ListProblems(ctx *gin.Context) {
	c.listProblems(ctx)
}

func (c *ProblemController) MineProblems(ctx *gin.Context) {
	c.allProblems(ctx, true)
}

func (c *ProblemController) AllProblems(ctx *gin.Context) {
	c.allProblems(ctx, false)
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

// 上传测试数据
func (c *ProblemController) UploadTestData(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.TestdataUploadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存上传文件
	uploadPath := os.TempDir() + "/upload_" + strconv.FormatUint(uint64(req.ProblemID), 10) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + ".zip"
	if err := ctx.SaveUploadedFile(file, uploadPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(uploadPath)

	// 处理测试数据
	if err := c.problemService.UploadTestdata(req.ProblemID, uploadPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// 下载测试数据
func (c *ProblemController) DownloadTestData(ctx *gin.Context) {
	var req model.TestdataDownloadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	zipPath, err := c.problemService.DownloadTestdata(req.ProblemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 设置响应头让浏览器下载文件
	ctx.FileAttachment(*zipPath, "problem_"+strconv.FormatUint(uint64(req.ProblemID), 10)+"_data.zip")

}

// 删除测试数据
func (c *ProblemController) DeleteTestData(ctx *gin.Context) {
	var req model.TestdataDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.problemService.DeleteTestdata(req.ProblemID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// 上传配置文件
func (c *ProblemController) UploadConfig(ctx *gin.Context) {
	var req model.TestdataConfigUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.problemService.UploadConfig(req.ProblemID, &req.Config); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// 获取配置文件
func (c *ProblemController) GetConfig(ctx *gin.Context) {
	var req model.TestdataConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config, err := c.problemService.GetConfig(req.ProblemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"config": config})
}
