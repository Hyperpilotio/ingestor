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
	Config  *viper.Viper
	mutex   sync.Mutex
	runLoop bool
}

// NewServer return an instance of Server struct.
func NewServer(config *viper.Viper) *Server {
	return &Server{
		Config: config,
	}
}

func (server *Server) runCaptureLoop(interval time.Duration, capturers *capturer.Capturers) {
	for server.runLoop {
		timer := time.NewTimer(interval)
		glog.Infof("Waiting for %s before moving to next capture", interval)
		err := capturers.Run()
		if err != nil {
			glog.Warningf("Error when running capturers: %s", err.Error())
		}
		<-timer.C
	}
}

// startCapture starts the capture loop if not already started
func (server *Server) startCapture() error {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	if server.runLoop {
		return fmt.Errorf("Ingestor already started")
	}

	interval, err := time.ParseDuration(server.Config.GetString("interval"))
	if err != nil {
		return fmt.Errorf("Unable to parse interval %s", interval, err.Error())
	}

	capturers, err := capturer.NewCapturers(server.Config)
	if err != nil {
		return fmt.Errorf("Unable to create capturers", err.Error())
	}

	server.runLoop = true
	go server.runCaptureLoop(interval, capturers)

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
	if err := server.startCapture(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
			"data":  "Unable to start ingestor: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"error": false,
	})
}

func (server *Server) stopIngestor(c *gin.Context) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if !server.runLoop {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
			"data":  "Ingestor already stopped",
		})
		return
	}

	server.runLoop = false

	c.JSON(http.StatusAccepted, gin.H{
		"error": false,
	})
}
