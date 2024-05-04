package routes

import (
	"github.com/draco121/common/constants"
	"github.com/draco121/common/middlewares"
	"github.com/draco121/common/utils"
	"github.com/gin-gonic/gin"
	"trainingservice/controllers"
)

func RegisterRoutes(controllers controllers.Controllers, router *gin.Engine) {
	utils.Logger.Info("Registering routes...")
	v1 := router.Group("/v1")
	// Register UploadTrainingData controller function
	v1.POST("/upload/:projectId/:botId", middlewares.AuthMiddleware(constants.Write), controllers.UploadTrainingData)

	// Register DeleteFile controller function
	v1.DELETE("/delete/:projectId/:botId/:fileId", middlewares.AuthMiddleware(constants.Write), controllers.DeleteFile)

	// Register GetFile controller function
	v1.GET("/download/:projectId/:botId/:fileId", middlewares.AuthMiddleware(constants.Read), controllers.GetFile)

	// Register AddTrainingData controller function
	v1.POST("/trainingdata", middlewares.AuthMiddleware(constants.Write), controllers.AddTrainingData)

	// Register GetTrainingData controller function
	v1.GET("/trainingdata", middlewares.AuthMiddleware(constants.Read), controllers.GetTrainingData)

	// Register UpdateTrainingData controller function
	v1.PATCH("/trainingdata", middlewares.AuthMiddleware(constants.Write), controllers.GetTrainingData)

	// Register ResetTrainingData controller function
	v1.DELETE("/trainingdata", middlewares.AuthMiddleware(constants.Write), controllers.DeleteTrainingData)

	utils.Logger.Info("Routes registered")
}
