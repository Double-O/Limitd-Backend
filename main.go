package main

import (
	"fmt"
	"net/http"

	"github.com/Double-O/Limitd-Backend/domain/entity"

	"github.com/Double-O/Limitd-Backend/middleware"
	"github.com/Double-O/Limitd-Backend/utils"

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

	// TODO : Change this level
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

	// login, refresh can be called by anyone, logout should be made by loggedin users
	authPublicRouter := r.Group("/auth")
	authPrivateRouter := r.Group("/auth")

	authPrivateRouter.Use(middleware.AuthMiddleware(userService, redisClient))

	authPublicRouter.POST("/login", controller.HandleLogin(userService, redisClient))
	authPublicRouter.GET("/refresh", controller.HandleRefresh(userService, redisClient))

	authPrivateRouter.POST("/logout", controller.HandleLogOut(redisClient))

	v1Router := r.Group("/v1")
	v1Router.Use(middleware.AuthMiddleware(userService, redisClient))

	// TODO need to be removed
	v1Router.GET("/ping", func(ctx *gin.Context) {
		userG, ok := ctx.Get(utils.USER)
		if !ok {
			logger.LogMessage(zerolog.ErrorLevel, "main.main", "main", "user koi vai")
			ctx.JSON(http.StatusOK, gin.H{
				"message": "no pong",
			})
			return
		}
		user := userG.(*entity.User)
		logger.LogMessage(zerolog.InfoLevel, "main.main", "main", fmt.Sprintf("user : %+v", user))
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}
