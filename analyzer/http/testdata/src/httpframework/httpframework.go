package httpframework

import (
	"net/http"

	"github.com/go-chi/chi" // want `third-party HTTP framework "github.com/go-chi/chi"; use the net/http standard library`
)

func Mux() *http.ServeMux {
	return http.NewServeMux()
}

func Router() *chi.Router {
	return chi.NewRouter()
}
