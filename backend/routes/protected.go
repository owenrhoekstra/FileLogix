package routes

import (
	"FileLogix/elevation"
	"FileLogix/internal/createRecord"
	"FileLogix/internal/editRecord"
	"FileLogix/internal/fileRecord"
	"FileLogix/internal/viewRecord"
	"FileLogix/middleware"

	"net/http"
)

func ProtectedRoutes() http.Handler {
	mux := http.NewServeMux()

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

	mux.Handle("GET /documents/{id}",
		middleware.RequirePermission("can_read")(
			http.HandlerFunc(viewRecord.FetchRecordDetails),
		),
	)

	mux.Handle("DELETE /documents/{id}",
		middleware.RequirePermission("can_delete")(
			elevation.RequireActionElevation(
				http.HandlerFunc(viewRecord.DeleteRecord),
			),
		),
	)

	mux.Handle("PATCH /documents/{id}",
		middleware.RequirePermission("can_edit")(
			http.HandlerFunc(editRecord.HandleRecordEdit),
		),
	)

	mux.Handle("/documents/{id}/restore",
		middleware.RequirePermission("can_delete")(
			http.HandlerFunc(editRecord.HandleRecordRestore),
		),
	)

	mux.Handle("/settings/",
		middleware.RequireRole("superuser", "manager")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.StripPrefix("/settings", SettingsRoutes()).ServeHTTP(w, r)
			}),
		),
	)

	return mux
}
