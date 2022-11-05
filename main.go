package main

import (
	"fmt"
	"net/http"

	"github.com/Double-O/Limitd-Backend/middleware"

	"github.com/go-redis/redis/v9"

	"github.com/Double-O/Limitd-Backend/controller"
	"github.com/Double-O/Limitd-Backend/initializers"

	"github.com/Double-O/Limitd-Backend/logger"
	"github.com/Double-O/Limitd-Backend/service"
	"github.com/rs/zerolog"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var db *gorm.DB
var redisClient *redis.Client
var userService service.UserService

func setUp() error {
	var err error

	//load env variable
	godotenv.Load()

	// set log level
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	//initializing db
	db, err = initializers.NewDbConnection()
	if err != nil {
		logger.LogMessage(zerolog.ErrorLevel, "main.main", "setUp", "Db connection failed")
		return err
	}

	//initializing redis
	redisClient = initializers.NewRedisClient()

	// initializing user service
	userService = service.NewUserService(db)
	return nil
}

func main() {
	fmt.Println("Yo")
	r := gin.Default()

	err := setUp()
	if err != nil {
		panic("setup failed")
	}

	authRouter := r.Group("/auth")
	authRouter.POST("/login", controller.HandleLogin(userService, redisClient))

	v1Router := r.Group("/v1")
	v1Router.Use(middleware.AuthMiddleware(redisClient))

	v1Router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}
