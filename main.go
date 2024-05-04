package main

import (
	"github.com/draco121/common/database"
	"github.com/draco121/common/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"trainingservice/controllers"
	"trainingservice/core"
	"trainingservice/repository"
	"trainingservice/routes"
)

func RunApp() {
	utils.Logger.Info("starting trainingservice...")
	client := database.NewMongoDatabase(os.Getenv("MONGODB_URI"))
	utils.Logger.Debug(utils.BaseDir())
	db := client.Database("training-service")
	repo := repository.NewTrainingRepository(db)
	service := core.NewTrainingService(client, repo)
	controller := controllers.NewControllers(service)
	router := gin.New()
	router.Use(gin.LoggerWithWriter(utils.Logger.Out))
	routes.RegisterRoutes(controller, router)
	utils.Logger.Info("started trainingservice...")
	err := router.Run()
	if err != nil {
		utils.Logger.Fatal(err.Error())
		return
	}
}
func main() {
	_ = godotenv.Load()
	RunApp()
}
