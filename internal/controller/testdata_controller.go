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

type TestdataController struct {
	testdataService *service.TestdataService
}

func NewTestdataController(testdataService *service.TestdataService) *TestdataController {
	return &TestdataController{
		testdataService: testdataService,
	}
}

// UploadTestData 上传测试数据
func (c *TestdataController) UploadTestData(ctx *gin.Context) {
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
	if err := c.testdataService.UploadTestdata(req.ProblemID, uploadPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DownloadTestData 下载测试数据
func (c *TestdataController) DownloadTestData(ctx *gin.Context) {
	var req model.TestdataDownloadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zipPath, err := c.testdataService.DownloadTestdata(req.ProblemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头让浏览器下载文件
	ctx.FileAttachment(zipPath, "problem_"+strconv.FormatUint(uint64(req.ProblemID), 10)+"_data.zip")

	// 文件发送完成后删除临时文件
	go func() {
		time.Sleep(30 * time.Second) // 等待足够时间确保文件已发送
		os.Remove(zipPath)
	}()
}

// DeleteTestData 删除测试数据
func (c *TestdataController) DeleteTestData(ctx *gin.Context) {
	var req model.TestdataDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.testdataService.DeleteTestdata(req.ProblemID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// UploadConfig 上传配置文件
func (c *TestdataController) UploadConfig(ctx *gin.Context) {
	var req model.TestdataConfigUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.testdataService.UploadConfig(req.ProblemID, &req.Config); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetConfig 获取配置文件
func (c *TestdataController) GetConfig(ctx *gin.Context) {
	var req model.TestdataConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := c.testdataService.GetConfig(req.ProblemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"config": config})
}