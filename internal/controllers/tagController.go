package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unbot2313/go-streaming-service/internal/helpers"
	"github.com/unbot2313/go-streaming-service/internal/models"
	"github.com/unbot2313/go-streaming-service/internal/services"
)

type TagControllerImpl struct {
	tagService           services.TagService
	databaseVideoService services.DatabaseVideoService
}

type TagController interface {
	GetAllTags(c *gin.Context)
	GetVideosByTag(c *gin.Context)
	AddTagsToVideo(c *gin.Context)
	RemoveTagFromVideo(c *gin.Context)
}

func NewTagController(tagService services.TagService, databaseVideoService services.DatabaseVideoService) TagController {
	return &TagControllerImpl{
		tagService:           tagService,
		databaseVideoService: databaseVideoService,
	}
}

type AddTagsRequest struct {
	Tags []string `json:"tags" binding:"required,min=1,max=10"`
}

type RemoveTagRequest struct {
	Tag string `json:"tag" binding:"required"`
}

// GetAllTags godoc
// @Summary		Get all tags
// @Description	Retrieve all available tags sorted alphabetically
// @Tags		tags
// @Produce		json
// @Success		200 {object} helpers.APIResponse{data=[]models.Tag}
// @Failure		500 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/tags [get]
func (tc *TagControllerImpl) GetAllTags(c *gin.Context) {
	tags, err := tc.tagService.GetAllTags()
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not retrieve tags", err)
		return
	}

	helpers.Success(c, http.StatusOK, tags)
}

// GetVideosByTag godoc
// @Summary		Get videos by tag
// @Description	Retrieve videos that have a specific tag. Supports pagination.
// @Tags		tags
// @Produce		json
// @Param		tag path string true "Tag name"
// @Param		page query int false "Page number (default: 1)" default(1)
// @Param		page_size query int false "Items per page (default: 10, max: 50)" default(10)
// @Success		200 {object} helpers.APIResponse{data=services.PaginatedVideos}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		500 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/tags/{tag}/videos [get]
func (tc *TagControllerImpl) GetVideosByTag(c *gin.Context) {
	tagName := c.Param("tag")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	result, err := tc.tagService.FindVideosByTag(tagName, page, pageSize)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Tag not found", err)
		return
	}

	helpers.Success(c, http.StatusOK, result)
}

// AddTagsToVideo godoc
// @Summary		Add tags to a video
// @Description	Add one or more tags to a video. Only the video owner can add tags. Tags are created if they don't exist.
// @Tags		tags
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		videoid path string true "Video ID"
// @Param		body body AddTagsRequest true "Tags to add"
// @Success		200 {object} helpers.APIResponse{data=object{message=string}}
// @Failure		400 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		403 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/tags/{videoid} [post]
func (tc *TagControllerImpl) AddTagsToVideo(c *gin.Context) {
	videoId := c.Param("videoid")

	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	authenticatedUser := user.(*models.User)

	video, err := tc.databaseVideoService.FindVideoByID(videoId)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Video not found", err)
		return
	}

	if video.UserID != authenticatedUser.Id {
		helpers.HandleError(c, http.StatusForbidden, "You are not the owner of this video", nil)
		return
	}

	var req AddTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "tags field is required (array of strings, max 10)", err)
		return
	}

	if err := tc.tagService.AddTagsToVideo(videoId, req.Tags); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "Could not add tags", err)
		return
	}

	helpers.Success(c, http.StatusOK, gin.H{"message": "Tags added successfully"})
}

// RemoveTagFromVideo godoc
// @Summary		Remove a tag from a video
// @Description	Remove a specific tag from a video. Only the video owner can remove tags.
// @Tags		tags
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		videoid path string true "Video ID"
// @Param		body body RemoveTagRequest true "Tag to remove"
// @Success		200 {object} helpers.APIResponse{data=object{message=string}}
// @Failure		400 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		401 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		403 {object} helpers.APIResponse{error=helpers.APIError}
// @Failure		404 {object} helpers.APIResponse{error=helpers.APIError}
// @Router		/tags/{videoid} [delete]
func (tc *TagControllerImpl) RemoveTagFromVideo(c *gin.Context) {
	videoId := c.Param("videoid")

	user, exists := c.Get("user")
	if !exists {
		helpers.HandleError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	authenticatedUser := user.(*models.User)

	video, err := tc.databaseVideoService.FindVideoByID(videoId)
	if err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Video not found", err)
		return
	}

	if video.UserID != authenticatedUser.Id {
		helpers.HandleError(c, http.StatusForbidden, "You are not the owner of this video", nil)
		return
	}

	var req RemoveTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "tag field is required", err)
		return
	}

	if err := tc.tagService.RemoveTagFromVideo(videoId, req.Tag); err != nil {
		helpers.HandleError(c, http.StatusNotFound, "Tag not found", err)
		return
	}

	helpers.Success(c, http.StatusOK, gin.H{"message": "Tag removed successfully"})
}
