package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/h3ll0kitt1/avitotest/internal/models"
)

func (app *application) getHistory(w http.ResponseWriter, r *http.Request) {

	historyDownload := struct {
		Users []int64 `json:"user_list"`
		Days  int     `json:"days"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&historyDownload)
	if err != nil {
		errorWrongFormat(w)
		return
	}

	for _, user := range historyDownload.Users {

		ok := app.validator.UserId(user)
		if !ok {
			errorWrongFormat(w)
			return
		}
	}

	ok := app.validator.Days(historyDownload.Days)
	if !ok {
		errorWrongFormat(w)
		return
	}

	history, err := app.storage.GetHistory(r.Context(), historyDownload.Users, historyDownload.Days)
	if err != nil {
		errorInternalServer(w)
		return
	}

	filename, err := app.file.Download(history)
	if err != nil {
		errorInternalServer(w)
		return
	}

	jsonData, err := json.Marshal(filename)
	if err != nil {
		errorInternalServer(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

func (app *application) createSegment(w http.ResponseWriter, r *http.Request) {

	slug := chi.URLParam(r, "slug")
	ok := app.validator.SegmentSlug(slug)
	if !ok {
		errorWrongFormat(w)
		return
	}

	segment := struct {
		PercentageRND int `json:"percentage_random"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&segment)
	if err != nil {
		errorWrongFormat(w)
		return
	}

	ok = app.validator.PercentageRND(segment.PercentageRND)
	if !ok {
		errorWrongFormat(w)
		return
	}

	if err := app.storage.CreateSegment(r.Context(), slug, segment.PercentageRND); err != nil {
		errorInternalServer(w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) deleteSegment(w http.ResponseWriter, r *http.Request) {

	slug := chi.URLParam(r, "slug")
	ok := app.validator.SegmentSlug(slug)
	if !ok {
		errorWrongFormat(w)
		return
	}

	if err := app.storage.DeleteSegment(r.Context(), slug); err != nil {
		errorInternalServer(w)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) getSegments(w http.ResponseWriter, r *http.Request) {

	userStr := chi.URLParam(r, "user_id")

	user, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		errorWrongFormat(w)
		return
	}

	ok := app.validator.UserId(user)
	if !ok {
		errorWrongFormat(w)
		return
	}

	segments, err := app.storage.GetSegmentsByUserID(r.Context(), user)
	if err != nil {
		errorInternalServer(w)
		return
	}

	jsonData, err := json.Marshal(segments)
	if err != nil {
		errorInternalServer(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

func (app *application) updateSegments(w http.ResponseWriter, r *http.Request) {

	userStr := chi.URLParam(r, "user_id")
	user, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		errorWrongFormat(w)
		return
	}

	ok := app.validator.UserId(user)
	if !ok {
		errorWrongFormat(w)
		return
	}

	segments := struct {
		Delete []models.Segment `json:"list_delete,omitempty"`
		Add    []models.Segment `json:"list_add,omitempty"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&segments)
	if err != nil {
		errorWrongFormat(w)
		return
	}

	ok = app.validator.Segments(segments.Delete)
	if !ok {
		errorWrongFormat(w)
		return
	}

	ok = app.validator.Segments(segments.Add)
	if !ok {
		errorWrongFormat(w)
		return
	}

	if err := app.storage.UpdateSegmentsByUserID(r.Context(), user, segments.Delete, segments.Add); err != nil {
		errorInternalServer(w)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func errorNotFound(w http.ResponseWriter, r *http.Request) {
	message := []byte(`{"error": { "code": 404, "message": "Wrong resource url"} }`)
	w.WriteHeader(http.StatusNotFound)
	w.Write(message)
}

func errorInternalServer(w http.ResponseWriter) {
	message := []byte(`{"error": { "code": 500, "message": "Error while processing request. Please, contact support"} }`)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(message)
}

func errorWrongFormat(w http.ResponseWriter) {
	message := []byte(`{"error": { "code": 400, "message": "Wrong body request format"} }`)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(message)
}
