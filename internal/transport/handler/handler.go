package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/RomanKovalev007/organization_service/internal/apperr"
	"github.com/RomanKovalev007/organization_service/internal/domain"
	"github.com/RomanKovalev007/organization_service/internal/service"
)

type Handler struct {
	svc departmentService
}

func New(svc departmentService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "name is required")
		return
	}
	if len(req.Name) > 200 {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "name must be at most 200 characters")
		return
	}

	dept, err := h.svc.CreateDepartment(r.Context(), &domain.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		handleAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dept)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid department id")
		return
	}

	var req CreateEmployeeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid request body")
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Position = strings.TrimSpace(req.Position)

	if req.FullName == "" {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "full_name is required")
		return
	}
	if len(req.FullName) > 200 {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "full_name must be at most 200 characters")
		return
	}
	if req.Position == "" {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "position is required")
		return
	}
	if len(req.Position) > 200 {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "position must be at most 200 characters")
		return
	}

	emp, err := h.svc.CreateEmployee(r.Context(), &domain.Employee{
		DepartmentID: id,
		FullName:     req.FullName,
		Position:     req.Position,
		HiredAt:      req.HiredAt,
	})
	if err != nil {
		handleAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, emp)
}

func (h *Handler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid department id")
		return
	}

	depth := 1
	if raw := r.URL.Query().Get("depth"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 5 {
			writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "depth must be between 1 and 5")
			return
		}
		depth = parsed
	}
	includeEmployees := r.URL.Query().Get("include_employees") == "true"

	tree, err := h.svc.GetDepartment(r.Context(), id, depth, includeEmployees)
	if err != nil {
		handleAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tree)
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid department id")
		return
	}

	var req UpdateDepartmentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid request body")
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		req.Name = &trimmed
		if trimmed == "" {
			writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "name cannot be empty")
			return
		}
		if len(trimmed) > 200 {
			writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "name must be at most 200 characters")
			return
		}
	}

	dept, err := h.svc.UpdateDepartment(r.Context(), id, req.Name, req.ParentID)
	if err != nil {
		handleAppErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dept)
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid department id")
		return
	}

	mode := service.DeleteMode(r.URL.Query().Get("mode"))
	if mode != service.DeleteModeCascade && mode != service.DeleteModeReassign {
		writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "mode must be 'cascade' or 'reassign'")
		return
	}

	var reassignTo *int64
	if raw := r.URL.Query().Get("reassign_to_department_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, apperr.CodeInvalidInput, "invalid reassign_to_department_id")
			return
		}
		reassignTo = &parsed
	}

	if err := h.svc.DeleteDepartment(r.Context(), id, mode, reassignTo); err != nil {
		handleAppErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
