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
	userService := service.NewUserService(userRepo)

	problemService := service.NewProblemService(problemRepo)
	testdataService := service.NewTestdataService(problemRepo, "/var/testdata")
	judgeService := service.NewJudgeService(submissionRepo, problemRepo, userRepo, "/var/testdata")
	fileService := service.NewFileService("/var/www/reisen/file")

	// Initialize controllers
	configController := controller.NewConfigController()
	problemController := controller.NewProblemController(problemService)
	recordController := controller.NewRecordController(judgeService)
	testdataController := controller.NewTestdataController(testdataService)
	avatarController := controller.NewAvatarController(fileService, userRepo)
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)

	// Initialize router
	router := gin.Default()

	// Public routes
	public := router.Group("/api")
	{
		public.GET("/sync-config", configController.SyncConfig)

		public.POST("/auth/login", authController.Login)
		public.POST("/auth/register", authController.Register)

    public.POST("/problem", problemController.GetProblem)
		public.POST("/problem/list", problemController.ListProblemsPublic)

    public.POST("/record", recordController.GetRecordDetail)
    public.POST("/record/list", recordController.ListRecords)

    public.POST("/user", userController.GetUser)
	}

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.POST("/auth/me", authController.Me)
		protected.POST("/auth/logout", authController.Logout)
		protected.POST("/auth/reset", authController.SetPassword)

		protected.POST("/user/edit", userController.EditUser)
		protected.POST("/user/delete", userController.DeleteUser)

    protected.POST("/problem/submit", recordController.SubmitCode)
		protected.POST("/problem/mine", problemController.ListProblemsMine)

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
			adminRoutes.POST("/user/all", userController.ListUsers)
			adminRoutes.POST("/problem/all", problemController.ListProblemsAdmin)
    	adminRoutes.POST("/record/all", recordController.ListRecords)
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