package routes

import (
	"FileLogix/elevation"
	"FileLogix/internal/createRecord"
	"FileLogix/internal/editRecord"
	"FileLogix/internal/fileRecord"
	"FileLogix/internal/viewRecord"
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
			http.HandlerFunc(createRecord.Create),
		),
	)

	mux.Handle("/form-metadata",
		middleware.RequirePermission("can_read")(
			http.HandlerFunc(createRecord.TypeOptions),
		),
	)

	mux.Handle("/print/{id}",
		middleware.RequirePermission("can_write")(
			http.HandlerFunc(createRecord.PrintLabel),
		),
	)

	mux.Handle("/records/location",
		middleware.RequirePermission("can_file")(
			http.HandlerFunc(fileRecord.File),
		),
	)

	mux.Handle("/cabinets/{id}",
		middleware.RequirePermission("can_read")(
			http.HandlerFunc(fileRecord.CabinetMeta),
		),
	)

	mux.Handle("/fetch-records",
		middleware.RequirePermission("can_read")(
			http.HandlerFunc(viewRecord.FetchRecordList),
		),
	)

	mux.Handle("/files/{path...}",
		middleware.RequirePermission("can_read")(
			http.HandlerFunc(viewRecord.FetchDocumentImages),
		),
	)

	mux.Handle("/documents/{id}",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				middleware.RequirePermission("can_read")(
					http.HandlerFunc(viewRecord.FetchRecordDetails),
				).ServeHTTP(w, r)

			case http.MethodDelete:
				middleware.RequirePermission("can_delete")(
					elevation.RequireActionElevation(
						http.HandlerFunc(viewRecord.DeleteRecord),
					),
				).ServeHTTP(w, r)

			case http.MethodPatch:
				middleware.RequirePermission("can_edit")(
					http.HandlerFunc(editRecord.HandleRecordEdit),
				).ServeHTTP(w, r)

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}),
	)

	mux.Handle("/documents/{id}/restore",
		middleware.RequirePermission("can_delete")(
			http.HandlerFunc(editRecord.HandleRecordRestore),
		),
	)

	return mux
}
