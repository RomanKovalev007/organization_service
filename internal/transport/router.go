package transport

import (
	"net/http"

	"github.com/RomanKovalev007/organization_service/internal/transport/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /departments/", h.CreateDepartment)
	mux.HandleFunc("POST /departments/{id}/employees/", h.CreateEmployee)
	mux.HandleFunc("GET /departments/{id}", h.GetDepartment)
	mux.HandleFunc("PATCH /departments/{id}", h.UpdateDepartment)
	mux.HandleFunc("DELETE /departments/{id}", h.DeleteDepartment)

	return mux
}
