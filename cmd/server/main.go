package main

import (
	"reisen-be/internal/config"
	"reisen-be/internal/controller"
	"reisen-be/internal/middleware"
	"reisen-be/internal/model"
	"reisen-be/internal/repository"
	"reisen-be/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate models
	if err := db.AutoMigrate(&model.User{}); err != nil {
		panic("failed to migrate database")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	problemService := service.NewProblemService(problemRepo)
	judgeService := service.NewJudgeService(submissionRepo, problemRepo, userRepo, "/var/testdata")
	testdataService := service.NewTestdataService(problemRepo, "/var/testdata")
	migrationService := service.NewGormMigrationService(db)

	// Migrate
	migrationService.RunMigrations()

	// Initialize controllers
	authController := controller.NewAuthController(authService)
	configController := controller.NewConfigController()
	problemController := controller.NewProblemController(problemService)
	recordController := controller.NewRecordController(judgeService)
	testdataController := controller.NewTestdataController(testdataService)

	// Initialize router
	router := gin.Default()

	// Public routes
	public := router.Group("/api")
	{
		public.GET("/sync-config", configController.SyncConfig)

		public.POST("/auth/login", authController.Login)
		public.POST("/auth/register", authController.Register)

    public.POST("/problem", problemController.GetProblem)
		public.POST("/problem/list", problemController.ListProblems)

    public.POST("/record", recordController.GetRecordDetail)
    public.POST("/record/list", recordController.ListRecords)

	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.POST("/logout", authController.Logout)
		protected.POST("/problem/edit", problemController.CreateOrUpdateProblem)
    protected.POST("/problem/submit", recordController.SubmitCode)
    protected.POST("/problem/delete", problemController.DeleteProblem)

		protected.POST("/testdata", testdataController.DownloadTestData)
		protected.POST("/testdata/upload", testdataController.UploadTestData)
		protected.POST("/testdata/delete", testdataController.DeleteTestData)
		protected.POST("/testdata/config", testdataController.GetConfig)
		protected.POST("/testdata/config/upload", testdataController.UploadConfig)
	}

	// Start server
	router.Run(":" + cfg.Server.Port)
}