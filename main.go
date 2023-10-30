package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var s3_client *s3.S3

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func getS3Client() *s3.S3 {
	/* Environment variables needed AWS client connectivity are:
	 *    1. AWS_ACCESS_KEY_ID
	 *    2. AWS_SECRET_ACCESS_KEY
	 */
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	return s3.New(sess)
}

func sendResponse(level log.Level, fields log.Fields, msg string, c *gin.Context) {
	switch level {
	case log.InfoLevel:
		log.WithFields(fields).Info(msg)
		c.JSON(http.StatusCreated, gin.H{
			"msg": msg,
		})
		break
	case log.ErrorLevel:
		log.WithFields(fields).Error(msg)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": msg,
		})
		break
	case log.WarnLevel:
		log.WithFields(fields).Warn(msg)
		c.JSON(http.StatusAccepted, gin.H{
			"warn": msg,
		})
		break
	default:
		log.WithFields(fields).Info(msg)
		c.JSON(http.StatusCreated, gin.H{
			"msg": msg,
		})
		break
	}
}

func handleUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get form file.",
		})
		return
	}

	folder := c.DefaultQuery("folder", "")
	if folder == "" {
		folder = "stores"
	}

	bucket_name := os.Getenv("BUCKET_NAME")
	object_name := folder + "/" + file.Filename

	content, err := file.Open()
	if err != nil {
		sendResponse(log.ErrorLevel, log.Fields{
			"bucket_name": bucket_name,
			"file":        object_name,
		}, "Failed to open file", c)
		return
	}

	_, err = s3_client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket_name,
		Key:    &object_name,
		Body:   content,
	})
	if err != nil {
		sendResponse(log.ErrorLevel, log.Fields{
			"bucket_name": bucket_name,
			"file":        object_name,
		}, fmt.Sprintf("Failed to upload to S3: %v", err), c)
		return
	}

	sendResponse(log.InfoLevel, log.Fields{
		"bucket_name": bucket_name,
		"file":        object_name,
	}, "File successfully uploaded to S3", c)
}

func main() {
	s3_client = getS3Client()
	r := gin.Default()
	r.Use(gin.BasicAuth(gin.Accounts{
		os.Getenv("USERNAME"): os.Getenv("PASSWORD"),
	}))

	r.POST("/upload", handleUpload)
	port, ok := os.LookupEnv("PORT")
	if ok {
		r.Run(":" + port)
	} else {
		r.Run(":8080")
	}
}
