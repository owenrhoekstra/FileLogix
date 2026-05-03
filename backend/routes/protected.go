package routes

import (
	"FileLogix/elevation"
	"FileLogix/internal/records"
	"FileLogix/middleware"
	"FileLogix/ocr"

	"net/http"
)

func ProtectedRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/ocr",
		middleware.RequireRole("superuser", "manager", "user", "contributor")(
			elevation.RequireActionElevation(
				http.HandlerFunc(ocr.OcrEndpoint),
			),
		),
	)

	mux.Handle("/records",
		middleware.RequirePermission("can_write")(
			http.HandlerFunc(records.Create),
		),
	)

	mux.Handle("/form-metadata",
		middleware.RequirePermission("can_write")(
			http.HandlerFunc(records.TypeOptions),
		),
	)

    mux.Handle("/print/{id}",
        middleware.RequirePermission("can_write")(
            http.HandlerFunc(records.PrintLabel),
        ),
    )

	return mux
}
