package routes

import (
	"net/http"

	"FileLogix/elevation"
	"FileLogix/settings/documentTypes"
)

func SettingsRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/add-document-type",
		elevation.RequireViewElevation(
			http.HandlerFunc(documentTypes.AddDocumentType),
		),
	)

	return mux
}
