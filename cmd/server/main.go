package main

import (
	"reisen-be/internal/config"
	"reisen-be/internal/controller"
	"reisen-be/internal/middleware"
	"reisen-be/internal/model"
	"reisen-be/internal/query"
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
	contestService := service.NewContestService(contestListQuery, contestRepo, problemRepo, signupRepo, userRepo, rankingRepo, judgementRepo)

	problemService := service.NewProblemService(problemListQuery, problemRepo)
	testdataService := service.NewTestdataService(problemRepo, "/var/testdata")
	judgeService := service.NewJudgeService(submissionRepo, judgementRepo, problemRepo, userRepo, "/var/testdata")
	fileService := service.NewFileService("/var/www/reisen/file")

	// Initialize controllers
	configController := controller.NewConfigController()
	problemController := controller.NewProblemController(problemService)
	submissionController := controller.NewSubmissionController(judgeService, userService)
	testdataController := controller.NewTestdataController(testdataService)
	avatarController := controller.NewAvatarController(fileService, userRepo)
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService, judgeService, contestService)
	contestController := controller.NewContestController(contestService, userService)

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
		protected.POST("/contest/ranking", contestController.GetRanking)
		protected.POST("/contest/ranklist", contestController.GetRanklist)
		protected.POST("/contest/problemset", contestController.GetContestProblems)

		protected.POST("/user/edit", userController.EditUser)
		protected.POST("/user/delete", userController.DeleteUser)

		protected.POST("/problem/submit", submissionController.SubmitCode)
		protected.POST("/problem/mine", problemController.MineProblems)

		protected.POST("/avatar/upload", avatarController.UploadAvatar)

		juryRoutes := protected.Group("")
		juryRoutes.Use(middleware.RoleRequired(model.RoleJury))
		{
			juryRoutes.POST("/problem/edit", problemController.CreateOrUpdateProblem)
			juryRoutes.POST("/problem/delete", problemController.DeleteProblem)

			juryRoutes.POST("/testdata/upload", testdataController.UploadTestData)
			juryRoutes.POST("/testdata/delete", testdataController.DeleteTestData)
			juryRoutes.POST("/testdata/config/upload", testdataController.UploadConfig)
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
