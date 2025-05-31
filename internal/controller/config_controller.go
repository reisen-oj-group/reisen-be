package controller

import (
    "reisen-be/configs"

    "github.com/gin-gonic/gin"
)

type ConfigController struct{}

func NewConfigController() *ConfigController {
    return &ConfigController{}
}

// 同步配置文件
func (c *ConfigController) SyncConfig(ctx *gin.Context) {
    // 直接从配置中获取数据
    response := configs.SystemConfig
    
    // 如果需要从数据库获取，可以在这里添加数据库查询逻辑
    // 例如：
    // tags, err := c.tagRepo.GetAll()
    // if err != nil {
    //     ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    //     return
    // }
    // response.Tags = tags
    
    ctx.JSON(200, response)
}