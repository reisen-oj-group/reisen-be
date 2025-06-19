package main

import (
	"reisen-be/internal/config"
	"reisen-be/internal/controller"
	"reisen-be/internal/filesystem"
	"reisen-be/internal/middleware"
	"reisen-be/internal/model"
	"reisen-be/internal/query"
	"reisen-be/internal/repository"
	"reisen-be/internal/service"
	"reisen-be/internal/websocket"
	"time"

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
	// Initialize websockets
	submissionWs := websocket.NewSubmissionWs(100 * time.Millisecond)

	// Initialize filesystems
	problemFilesystem := filesystem.NewProblemFilesystem("/var/problemset")
	imageFilesystem := filesystem.NewImageFilesystem("/var/www/reisen/uploads/images")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	signupRepo := repository.NewSignupRepository(db)
	rankingRepo := repository.NewRankingRepository(db)
	contestRepo := repository.NewContestRepository(db)
	judgementRepo := repository.NewJudgementRepository(db)

	// Initialize queries
	problemListQuery := query.NewProblemListQuery(db)
	contestListQuery := query.NewContestListQuery(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	userService := service.NewUserService(userRepo)
	contestService := service.NewContestService(contestListQuery, contestRepo, problemRepo, submissionRepo, signupRepo, userRepo, rankingRepo, 5 * time.Second)

	imageService := service.NewImageService(userRepo, imageFilesystem)

	// 题库管理
	problemService := service.NewProblemService(
		problemListQuery,  // 查询题目列表
		problemRepo,       // 题目信息仓库
		problemFilesystem, // 题目数据管理
	)

	// 评测管理
	judgeService := service.NewJudgeService(
		submissionRepo,	// 提交记录仓库（管理评测详情）
		judgementRepo,  // 题目结果仓库（管理试题通过情况）
		problemRepo,    // 题目信息仓库（管理题目基本信息）
		userRepo,       // 用户仓库（管理提交者）
		problemFilesystem,
		submissionWs,
		contestService,
		5, // 评测机 worker 个数
	)

	// Initialize controllers
	configController := controller.NewConfigController()
	problemController := controller.NewProblemController(problemService, judgeService)
	submissionController := controller.NewSubmissionController(judgeService, userService, submissionWs)
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService, judgeService, contestService)
	contestController := controller.NewContestController(contestService, problemService, userService, judgeService)
	imageController := controller.NewImageController(imageService)

	// Initialize router
	router := gin.Default()

	// Public routes
	public := router.Group("/api")
	publicOptional := public.Group("")
	publicOptional.Use(middleware.AuthMiddleware(authService, false))
	{
		public.GET("/sync-config", configController.SyncConfig)

		public.POST("/auth/login", authController.Login)
		public.POST("/auth/register", authController.Register)

		publicOptional.POST("/problem", problemController.GetProblem)
		publicOptional.POST("/problem/list", problemController.ListProblems)

		publicOptional.POST("/contest", contestController.GetContest)
		publicOptional.POST("/contest/list", contestController.ListContests)

		public.GET("/ws/submission/:id", submissionController.HandleSubmissionWS)
		public.POST("/submission", submissionController.GetSubmissionDetail)
		public.POST("/submission/list", submissionController.ListSubmissions)

		public.POST("/user", userController.GetUser)
		public.POST("/user/practice", userController.GetPractice)
	}

	// Protected routes, must auth
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(authService, true))
	{
		protected.POST("/auth/me", authController.Me)
		protected.POST("/auth/logout", authController.Logout)
		protected.POST("/auth/reset", authController.SetPassword)

		protected.POST("/contest/signup", contestController.SignupContest)
		protected.POST("/contest/signout", contestController.SignoutContest)
		protected.POST("/contest/submit", contestController.SubmitCode)
		protected.POST("/contest/ranking", contestController.GetRanking)
		protected.POST("/contest/ranklist", contestController.GetRanklist)
		protected.POST("/contest/problemset", contestController.GetContestProblems)

		protected.POST("/user/edit", userController.EditUser)
		protected.POST("/user/delete", userController.DeleteUser)

		protected.POST("/problem/submit", problemController.SubmitCode)
		protected.POST("/problem/mine", problemController.MineProblems)

		protected.POST("/upload/avatar", imageController.UploadAvatar)

		juryRoutes := protected.Group("")
		juryRoutes.Use(middleware.RoleRequired(model.RoleJury))
		{
			juryRoutes.POST("/problem/edit", problemController.CreateOrUpdateProblem)
			juryRoutes.POST("/problem/delete", problemController.DeleteProblem)

			juryRoutes.POST("/contest/edit", contestController.CreateOrUpdateContest)
			juryRoutes.POST("/contest/delete", contestController.DeleteContest)

			juryRoutes.POST("/upload/banner", imageController.UploadBanner)

			juryRoutes.POST("/testdata/upload", problemController.UploadTestData)
			juryRoutes.POST("/testdata/download", problemController.DownloadTestData)
			juryRoutes.POST("/testdata/delete", problemController.DeleteTestData)
			juryRoutes.POST("/testdata/config/upload", problemController.UploadConfig)
		}

		adminRoutes := protected.Group("")
		adminRoutes.Use(middleware.RoleRequired(model.RoleAdmin))
		{
			adminRoutes.POST("/user/all", userController.AllUsers)
			adminRoutes.POST("/problem/all", problemController.AllProblems)
			adminRoutes.POST("/contest/all", contestController.AllContests)
			adminRoutes.POST("/submission/all", submissionController.AllSubmissions)
		}

		superRoutes := protected.Group("")
		superRoutes.Use(middleware.RoleRequired(model.RoleSuper))
		{
			superRoutes.POST("/auth/create", authController.Create)
		}
	}

	// Start server
	router.Run(":" + cfg.Server.Port)
}
