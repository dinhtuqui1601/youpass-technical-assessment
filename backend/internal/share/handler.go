package share

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/youpass/backend/pkg/response"
)

// Handler exposes the share-link HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler builds a share Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create godoc
// @Summary      Create a share link
// @Description  Creates a short link for a submission owned by the authenticated user. Returns the existing ACTIVE link when one already exists.
// @Tags         share
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Submission ID"
// @Success      201  {object}  response.Envelope{data=ShareLinkDTO}
// @Failure      400  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      403  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/submissions/{id}/share [post]
func (h *Handler) Create(c *gin.Context) {
	submissionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationError, "Invalid submission id")
		return
	}
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	dto, err := h.svc.Create(c.Request.Context(), ownerID, submissionID)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, dto)
}

// List godoc
// @Summary      List share links
// @Description  Lists the authenticated user's share links.
// @Tags         share
// @Security     BearerAuth
// @Produce      json
// @Param        limit   query     int  false  "Limit"   default(20)
// @Param        offset  query     int  false  "Offset"  default(0)
// @Success      200     {object}  response.Envelope{data=[]ShareLinkDTO}
// @Failure      401     {object}  response.Envelope
// @Router       /api/share-links [get]
func (h *Handler) List(c *gin.Context) {
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	out, err := h.svc.List(c.Request.Context(), ownerID, limit, offset)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, out)
}

// Get godoc
// @Summary      Get a share link
// @Description  Returns the owner-facing detail of a share link.
// @Tags         share
// @Security     BearerAuth
// @Produce      json
// @Param        code  path      string  true  "Short code"
// @Success      200   {object}  response.Envelope{data=ShareLinkDTO}
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Router       /api/share-links/{code} [get]
func (h *Handler) Get(c *gin.Context) {
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	dto, err := h.svc.Get(c.Request.Context(), ownerID, c.Param("code"))
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, dto)
}

// Update godoc
// @Summary      Enable/disable a share link
// @Description  Toggles a share link between ACTIVE and DISABLED. Deleted links cannot be reactivated.
// @Tags         share
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        code     path      string                 true  "Short code"
// @Param        request  body      UpdateShareLinkRequest true  "New status"
// @Success      200      {object}  response.Envelope{data=ShareLinkDTO}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      403      {object}  response.Envelope
// @Failure      404      {object}  response.Envelope
// @Router       /api/share-links/{code} [patch]
func (h *Handler) Update(c *gin.Context) {
	var req UpdateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeValidationError, err.Error())
		return
	}
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	dto, err := h.svc.UpdateStatus(c.Request.Context(), ownerID, c.Param("code"), Status(req.Status))
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, dto)
}

// Delete godoc
// @Summary      Delete a share link
// @Description  Soft-deletes a share link. The short code is never reused.
// @Tags         share
// @Security     BearerAuth
// @Produce      json
// @Param        code  path      string  true  "Short code"
// @Success      200   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Router       /api/share-links/{code} [delete]
func (h *Handler) Delete(c *gin.Context) {
	ownerID, err := primitive.ObjectIDFromHex(response.UserID(c))
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeInvalidAccessToken, "Invalid token subject")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), ownerID, c.Param("code")); err != nil {
		response.FailErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"deleted": true})
}

// Redirect godoc
// @Summary      Open a short link
// @Description  Resolves a short code and redirects to the shared submission page (HTTP 302). Not wrapped in the JSON envelope.
// @Tags         share
// @Param        code  path  string  true  "Short code"
// @Success      302
// @Failure      403
// @Failure      404
// @Router       /s/{code} [get]
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if !ValidCodeFormat(code) {
		response.Fail(c, http.StatusNotFound, response.CodeShareLinkNotFound, "Invalid short code")
		return
	}

	if _, err := h.svc.ResolveRedirect(c.Request.Context(), code); err != nil {
		response.FailErr(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusFound, "/shared/"+code)
}

// PublicShare godoc
// @Summary      Get shared submission
// @Description  Returns the submission content for a publicly accessible share link (authoritative access check).
// @Tags         share
// @Produce      json
// @Param        code  path      string  true  "Short code"
// @Success      200   {object}  response.Envelope{data=PublicShareDTO}
// @Failure      403   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Router       /api/public/shares/{code} [get]
func (h *Handler) PublicShare(c *gin.Context) {
	code := c.Param("code")
	dto, err := h.svc.PublicShare(c.Request.Context(), code)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	response.JSON(c, http.StatusOK, dto)
}

