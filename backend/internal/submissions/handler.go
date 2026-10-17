package submissions

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/pkg/response"
)

// Handler exposes the minimal submission endpoints needed for the demo.
type Handler struct {
	repo *Repository
}

// NewHandler builds a submissions Handler.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// CreateRequest is the payload for creating a submission.
type CreateRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required"`
}

// Create godoc
// @Summary      Create a submission
// @Description  Creates a submission owned by the authenticated user (demo module).
// @Tags         submissions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      CreateRequest  true  "Submission payload"
// @Success      201      {object}  response.Envelope{data=Submission}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Router       /api/submissions [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationError, err.Error())
		return
	}

	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	now := time.Now().UTC()
	sub := &Submission{
		OwnerID:   ownerID,
		Title:     req.Title,
		Content:   req.Content,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.repo.Create(c.Request.Context(), sub); err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, sub)
}

// List godoc
// @Summary      List my submissions
// @Description  Lists the submissions owned by the authenticated user, newest first.
// @Tags         submissions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Envelope{data=[]Submission}
// @Failure      401  {object}  response.Envelope
// @Router       /api/submissions [get]
func (h *Handler) List(c *gin.Context) {
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	subs, err := h.repo.FindByOwner(c.Request.Context(), ownerID)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	if subs == nil {
		subs = []*Submission{}
	}
	response.JSON(c, http.StatusOK, subs)
}
