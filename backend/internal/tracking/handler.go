package tracking

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/pkg/response"
)

// Handler exposes view-tracking endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds a tracking Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// PostViewRequest is the payload for recording a view event.
type PostViewRequest struct {
	ViewID string `json:"viewId" binding:"required"`
}

// StatsDTO is the response of the stats endpoint.
type StatsDTO struct {
	Code         string     `json:"code"`
	TotalViews   int64      `json:"totalViews"`
	LastViewedAt *time.Time `json:"lastViewedAt"`
}

// PostView godoc
// @Summary      Record a view
// @Description  Records a view event for a publicly accessible share link. The viewId is idempotent across retries.
// @Tags         tracking
// @Accept       json
// @Produce      json
// @Param        code     path      string           true  "Short code"
// @Param        request  body      PostViewRequest  true  "View event payload"
// @Success      202      {object}  response.Envelope
// @Failure      400      {object}  response.Envelope
// @Failure      403      {object}  response.Envelope
// @Failure      404      {object}  response.Envelope
// @Failure      503      {object}  response.Envelope
// @Router       /api/public/shares/{code}/views [post]
func (h *Handler) PostView(c *gin.Context) {
	var req PostViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationError, err.Error())
		return
	}

	if err := h.svc.RecordView(c.Request.Context(), c.Param("code"), req.ViewID, c.ClientIP(), c.Request.UserAgent()); err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusAccepted, gin.H{"accepted": true})
}

// GetStats godoc
// @Summary      Get share-link statistics
// @Description  Returns view statistics for a share link owned by the authenticated user.
// @Tags         tracking
// @Security     BearerAuth
// @Produce      json
// @Param        code  path      string  true  "Short code"
// @Success      200   {object}  response.Envelope{data=StatsDTO}
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Router       /api/share-links/{code}/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	code := c.Param("code")
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	stats, err := h.svc.Stats(c.Request.Context(), ownerID, code)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, StatsDTO{
		Code:         code,
		TotalViews:   stats.TotalViews,
		LastViewedAt: stats.LastViewedAt,
	})
}
