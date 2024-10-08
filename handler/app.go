package handler

import (
	docs "fp2/docs"
	"fp2/infra/postgres"
	"fp2/middleware"
	repository "fp2/repository/auth"
	commentRepository "fp2/repository/comment"
	photoRepository "fp2/repository/photo"
	smRepository "fp2/repository/social_media"
	userRepository "fp2/repository/users"
	services "fp2/services/auth"
	cServices "fp2/services/comment"
	pServices "fp2/services/photo"
	smServices "fp2/services/social_media"
	userServices "fp2/services/users"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
)

func StartApp() {
	port := os.Getenv("PORT")
	// Database
	db := postgres.GetDbInstance()
	validate := validator.New()

	authRepository := repository.NewAuthRepositoryImpl(db)
	userRepository := userRepository.NewUserRepositoryImpl(db)
	photoRepository := photoRepository.NewPhotoRepositoryImpl(db)
	socialMediaRepository := smRepository.NewSocialMediaRepositoryImpl(db)
	commentRepostiory := commentRepository.NewCommentRepositoryImpl(db)

	authenticationService := services.NewAuthServiceImpl(authRepository, validate)
	userService := userServices.NewUserServiceImpl(userRepository, authRepository, validate)
	socialMediaService := smServices.NewSocialMediaServiceImpl(socialMediaRepository, validate)
	photoService := pServices.NewPhotoServiceImpl(photoRepository, validate)
	commentService := cServices.NewCommentServiceImpl(commentRepostiory, photoRepository, validate)

	authenticationController := NewAuthenticationController(authenticationService)
	userController := NewUserController(userService)
	socialMediaController := NewSocialMediaController(socialMediaService)
	photoController := NewPhotoController(photoService)
	commentController := NewCommentController(commentService)

	service := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	service.GET("", func(ctx *gin.Context) {
		var endpoints = []string{ctx.Request.Host + "/socialmedias", ctx.Request.Host + "/photos", ctx.Request.Host + "/comments", ctx.Request.Host + "/users"}
		ctx.JSON(http.StatusOK, gin.H{
			"message":   "Hacktiv8 Golang KS08 - Final Project 2",
			"endpoints": endpoints,
		})
	})
	service.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	authenticationRouter := service.Group("/users")
	authenticationRouter.POST("/register", authenticationController.Register)
	authenticationRouter.POST("/login", authenticationController.Login)
	authenticationRouter.Use(middleware.AuthenticatedUser(userRepository))
	{
		authenticationRouter.PUT("", userController.UpdateUser)
		authenticationRouter.DELETE("", userController.DeleteUser)
	}
	socialMediaRouter := service.Group("/socialmedias")
	socialMediaRouter.Use(middleware.AuthenticatedUser(userRepository))
	{
		socialMediaRouter.POST("", socialMediaController.CreateSocialMedia)
		socialMediaRouter.GET("", socialMediaController.GetAllSocialMedia)
		socialMediaRouter.PUT("/:socialMediaId", middleware.AuthorizedUserSm(socialMediaRepository), socialMediaController.UpdateSocialMedia)
		socialMediaRouter.DELETE("/:socialMediaId", middleware.AuthorizedUserSm(socialMediaRepository), socialMediaController.DeleteSocialMedia)
	}
	photoRouter := service.Group("/photos")
	photoRouter.Use(middleware.AuthenticatedUser(userRepository))
	{
		photoRouter.POST("", photoController.CreatePhoto)
		photoRouter.GET("", photoController.GetAllPhoto)
		photoRouter.PUT("/:photoId", middleware.AuthorizedUserP(photoRepository), photoController.UpdatePhoto)
		photoRouter.DELETE("/:photoId", middleware.AuthorizedUserP(photoRepository), photoController.DeletePhoto)
	}
	commentRouter := service.Group("/comments")
	commentRouter.Use(middleware.AuthenticatedUser(userRepository))
	{
		commentRouter.POST("", commentController.CreateComment)
		commentRouter.GET("", commentController.GetAllComment)
		commentRouter.PUT("/:commentId", middleware.AuthorizedUserC(commentRepostiory), commentController.UpdateComment)
		commentRouter.DELETE("/:commentId", middleware.AuthorizedUserC(commentRepostiory), commentController.DeleteComment)
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: service,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
