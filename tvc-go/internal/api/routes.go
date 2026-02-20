package api

import (
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/projects", handleListProjects)
	mux.HandleFunc("POST /api/v1/projects", handleCreateProject)
	mux.HandleFunc("GET /api/v1/projects/{id}", handleGetProject)
	mux.HandleFunc("PUT /api/v1/projects/{id}", handleUpdateProject)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", handleDeleteProject)

	mux.HandleFunc("GET /api/v1/projects/{id}/traffic", handleListTraffic)
	mux.HandleFunc("GET /api/v1/projects/{id}/traffic/{logId}", handleGetTrafficLog)
	mux.HandleFunc("GET /api/v1/projects/{id}/traffic/stats", handleTrafficStats)

	mux.HandleFunc("GET /api/v1/projects/{id}/replays", handleListReplays)
	mux.HandleFunc("POST /api/v1/projects/{id}/replays", handleCreateReplay)
	mux.HandleFunc("GET /api/v1/projects/{id}/replays/{replayId}", handleGetReplay)
	mux.HandleFunc("POST /api/v1/projects/{id}/replays/{replayId}/start", handleStartReplay)
	mux.HandleFunc("POST /api/v1/projects/{id}/replays/{replayId}/stop", handleStopReplay)
	mux.HandleFunc("GET /api/v1/projects/{id}/replays/{replayId}/results", handleReplayResults)

	mux.HandleFunc("GET /api/v1/projects/{id}/schemas", handleListSchemas)
	mux.HandleFunc("POST /api/v1/projects/{id}/schemas", handleUploadSchema)
	mux.HandleFunc("POST /api/v1/projects/{id}/schemas/diff", handleSchemaDiff)

	return mux
}

func jsonResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(body)) //nolint:errcheck
}

func handleListProjects(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"projects":[],"message":"not yet implemented"}`)
}

func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleGetProject(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleListTraffic(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"traffic":[],"message":"not yet implemented"}`)
}

func handleGetTrafficLog(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleTrafficStats(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleListReplays(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"replays":[],"message":"not yet implemented"}`)
}

func handleCreateReplay(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleGetReplay(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleStartReplay(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleStopReplay(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleReplayResults(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleListSchemas(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"schemas":[],"message":"not yet implemented"}`)
}

func handleUploadSchema(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}

func handleSchemaDiff(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, `{"message":"not yet implemented"}`)
}
