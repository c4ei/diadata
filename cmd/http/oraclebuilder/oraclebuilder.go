package main

import (
	"time"

	//jwt "github.com/blockstatecom/gin-jwt"

	builderutils "oraclebuilder/utils"

	"github.com/99designs/keyring"
	models "github.com/diadata-org/diadata/pkg/model"
	"github.com/diadata-org/diadata/pkg/oraclebuilder"
	"github.com/diadata-org/diadata/pkg/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

const (
	cachingTime1Sec   = 1 * time.Second
	cachingTime20Secs = 20 * time.Second
	cachingTimeShort  = time.Minute * 2
	// cachingTimeMedium = time.Minute * 10
	cachingTimeLong = time.Minute * 100
)

var identityKey = "id"

func main() {

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	relStore, err := models.NewRelDataStore()
	if err != nil {
		log.Errorln("NewRelDataStore", err)
	}
	k8bridgeurl := utils.Getenv("K8SBRIDGE_URL", "127.0.0.1:50051")
	oraclebaseimage := utils.Getenv("ORACLE_BASE_IMAGE", "us.icr.io/dia-registry/oracles/oracle-baseimage:latest")
	oraclenamespace := utils.Getenv("ORACLE_NAMESPACE", "dia-oracle-feeder")

	ph := builderutils.NewPodHelper(oraclebaseimage, oraclenamespace)

	ring, _ := keyring.Open(keyring.Config{
		ServiceName:     "oraclebuilder",
		Server:          k8bridgeurl,
		AllowedBackends: []keyring.BackendType{keyring.K8Secret},
	})

	oracle := &oraclebuilder.Env{RelDB: relStore, PodHelper: ph, Keyring: ring}

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type,access-control-allow-origin, access-control-allow-headers"},
	}))
	routerGroup := r.Group("/oraclebuilder")

	routerGroup.POST("/create", oracle.Create)
	routerGroup.GET("/list", oracle.List)
	routerGroup.GET("/view", oracle.View)

	port := utils.Getenv("LISTEN_PORT", ":8080")

	executionMode := utils.Getenv("EXEC_MODE", "")
	if executionMode == "production" {
		err = r.Run(port)
		if err != nil {
			log.Error(err)
		}
	} else {
		err = r.Run(":8081")
		if err != nil {
			log.Error(err)
		}
	}

}
