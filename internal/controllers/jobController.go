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
// @Description	Get the status of a video processing job. Only the job owner can view it.
// @Tags		jobs
// @Produce		json
// @Security	BearerAuth
// @Param		jobid path string true "Job ID"
// @Success		200 {object} helpers.APIResponse{data=models.JobSwagger}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		403 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
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

	helpers.Success(c, http.StatusOK, job)
}
