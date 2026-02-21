package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/api/request"
	"github.com/tvc-org/tvc/internal/api/response"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/internal/storage"
	"github.com/tvc-org/tvc/pkg/logger"
)

type OrganizationHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewOrganizationHandler(store storage.Repository, log *logger.Logger) *OrganizationHandler {
	return &OrganizationHandler{store: store, log: log}
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrgRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if req.Name == "" || len(req.Name) > 255 {
		response.ValidationError(w, []response.FieldError{
			{Field: "name", Message: "Name is required (1-255 characters)"},
		})
		return
	}

	if req.Slug == "" {
		req.Slug = generateSlug(req.Name)
	}
	if !slugPattern.MatchString(req.Slug) {
		response.ValidationError(w, []response.FieldError{
			{Field: "slug", Message: "Slug must be lowercase alphanumeric with dashes"},
		})
		return
	}

	org := &models.Organization{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.store.CreateOrganization(r.Context(), org); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "An organization with this slug already exists")
			return
		}
		h.log.Error().Err(err).Msg("failed to create organization")
		response.InternalError(w)
		return
	}

	response.Created(w, org)
}

func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	org, err := h.store.GetOrganization(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Organization")
			return
		}
		h.log.Error().Err(err).Msg("failed to get organization")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": org,
	})
}

func (h *OrganizationHandler) List(w http.ResponseWriter, r *http.Request) {
	// For now, this would need the user's orgs from auth context
	// Placeholder: returns empty list until org membership is wired
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": []models.Organization{},
	})
}

type updateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	org, err := h.store.GetOrganization(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Organization")
			return
		}
		h.log.Error().Err(err).Msg("failed to get organization")
		response.InternalError(w)
		return
	}

	var req updateOrgRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Name != "" {
		if len(req.Name) > 255 {
			errs = append(errs, response.FieldError{Field: "name", Message: "Name must be 1-255 characters"})
		} else {
			org.Name = req.Name
		}
	}
	if req.Slug != "" {
		if !slugPattern.MatchString(req.Slug) {
			errs = append(errs, response.FieldError{Field: "slug", Message: "Slug must be lowercase alphanumeric with dashes"})
		} else {
			org.Slug = req.Slug
		}
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	if err := h.store.UpdateOrganization(r.Context(), org); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "An organization with this slug already exists")
			return
		}
		h.log.Error().Err(err).Msg("failed to update organization")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": org,
	})
}

func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := h.store.DeleteOrganization(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Organization")
			return
		}
		h.log.Error().Err(err).Msg("failed to delete organization")
		response.InternalError(w)
		return
	}

	response.NoContent(w)
}

// Member management

func (h *OrganizationHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	members, err := h.store.ListOrganizationMembers(r.Context(), orgID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list organization members")
		response.InternalError(w)
		return
	}

	if members == nil {
		members = []models.OrganizationMember{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": members,
	})
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func (h *OrganizationHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req addMemberRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.UserID == "" && req.Email == "" {
		errs = append(errs, response.FieldError{Field: "user_id", Message: "Either user_id or email is required"})
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		errs = append(errs, response.FieldError{Field: "role", Message: "Role must be admin, member, or viewer"})
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	// For now, support user_id only. Email-based invitation would need invite flow.
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(w, "Invalid user_id format")
		return
	}

	if err := h.store.AddOrganizationMember(r.Context(), orgID, userID, req.Role); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "User is already a member of this organization")
			return
		}
		h.log.Error().Err(err).Msg("failed to add organization member")
		response.InternalError(w)
		return
	}

	response.Created(w, map[string]interface{}{
		"message": "Member added successfully",
	})
}

func (h *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	userID, err := request.PathUUID(r, "userId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := h.store.RemoveOrganizationMember(r.Context(), orgID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Member")
			return
		}
		h.log.Error().Err(err).Msg("failed to remove organization member")
		response.InternalError(w)
		return
	}

	response.NoContent(w)
}
