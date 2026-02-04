package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/helpers"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type JobController interface {
	GetJobByID(c *gin.Context)
}

type JobControllerImpl struct {
	jobService services.JobService
}

func NewJobController(jobService services.JobService) JobController {
	return &JobControllerImpl{
		jobService: jobService,
	}
}

// GetJobByID godoc
// @Summary		Get job status by ID
// @Description	Get the status of a video processing job
// @Tags		jobs
// @Produce		json
// @Param		jobid path string true "Job ID"
// @Success		200 {object} models.JobSwagger{}
// @Failure		404 {object} map[string]string
// @Failure		500 {object} map[string]string
// @Router		/jobs/{jobid} [get]
func (jc *JobControllerImpl) GetJobByID(c *gin.Context) {
	jobId := c.Param("jobid")

	// Verificar usuario autenticado
	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	authenticatedUser, ok := user.(*models.User)
	if !ok {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not parse user data", nil)
		return
	}

	job, err := jc.jobService.GetJobByID(jobId)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Job not found", err)
		return
	}

	// Verificar que el job pertenece al usuario autenticado
	if job.UserID != authenticatedUser.Id {
		helpers.HandleError(c, http.StatusForbidden, "You can only view your own jobs", nil)
		return
	}

	c.JSON(http.StatusOK, job)
}
