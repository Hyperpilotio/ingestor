package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hyperpilotio/ingestor/capturer"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

var captureFlag bool = true

// Server store the stats / data of every deployment
type Server struct {
	Config      *viper.Viper
	mutex       sync.Mutex
	CaptureFlag bool
}

// NewServer return an instance of Server struct.
func NewServer(config *viper.Viper) *Server {
	return &Server{
		Config: config,
	}
}

// AutoIngestor activates capture in new goroutine at beginning
func (server *Server) AutoIngestor() error {
	interval, intervalErr := time.ParseDuration(server.Config.GetString("interval"))
	if intervalErr != nil {
		glog.Warningf("Unable to parse interval %s", interval, intervalErr.Error())
		return intervalErr
	}

	capturers, capturerErr := capturer.NewCapturers(server.Config)
	if capturerErr != nil {
		glog.Warningf("Unable to create capturers", capturerErr.Error())
		return capturerErr
	}

	server.CaptureFlag = true
	server.capture(interval, capturers)
	return nil
}

// StartServer start a web server
func (server *Server) StartServer() error {
	//gin.SetMode("release")
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	ingestorGroup := router.Group("/ingestor")
	{
		ingestorGroup.POST("/start", server.startIngestor)
		ingestorGroup.POST("/stop", server.stopIngestor)
	}

	return router.Run(":" + server.Config.GetString("port"))
}

func (server *Server) startIngestor(c *gin.Context) {
	if server.CaptureFlag {
		return
	}

	server.mutex.Lock()
	defer server.mutex.Unlock()

	interval, intervalErr := time.ParseDuration(server.Config.GetString("interval"))
	if intervalErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
			"data":  fmt.Sprintf("Unable to parse interval %s", interval, intervalErr.Error()),
		})
		return
	}

	capturers, capturerErr := capturer.NewCapturers(server.Config)
	if capturerErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": true,
			"data":  "Unable to create capturers: " + capturerErr.Error(),
		})
		return
	}

	// TODO: Only start if it hasn't been started!!!
	server.CaptureFlag = true
	server.capture(interval, capturers)
	c.JSON(http.StatusAccepted, gin.H{
		"error": false,
	})
}

func (server *Server) stopIngestor(c *gin.Context) {
	server.CaptureFlag = false
}

func (server *Server) capture(interval time.Duration, capturers *capturer.Capturers) {
	if server.CaptureFlag {
		timer := time.NewTimer(interval)
		glog.Infof("Waiting for %s before moving to next capture", interval)
		err := capturers.Run()
		if err != nil {
			glog.Warningf("Error when running capturers: %s", err.Error())
		}
		<-timer.C
		server.capture(interval, capturers)
	}
}
