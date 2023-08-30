package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/h3ll0kitt1/avitotest/internal/models"
)

// GetHistory godoc
//
//	@summary        Выгрузить историю
//	@description    Выгружает историю для переданных пользователей за переданное количество дней в csv файл и возвращает имя файла
//	@tags           history
//	@accept         json
//	@produce        json
//	@param          body    body    historyDownloadForm    true    "History form"
//	@success        200 string    string
//	@failure        400 {object}    errorResponse
//	@failure        500 {object}    errorResponse
//	@router         /history [get]
func (app *application) getHistory(w http.ResponseWriter, r *http.Request) {

	var form historyDownloadForm
	err := json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		app.logger.Errorw("error",
			"getHistory: error parsing historyDownloadForm", err,
		)
		app.errorWrongFormat(w)
		return
	}

	for _, user := range form.Users {

		ok := app.validator.UserId(user)
		if !ok {
			app.errorWrongFormat(w)
			return
		}
	}

	ok := app.validator.Days(form.Days)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	history, err := app.storage.GetHistory(r.Context(), form.Users, form.Days)
	if err != nil {
		app.logger.Errorw("error",
			"getHistory: error retrieving data from storage", err,
		)
		app.errorInternalServer(w)
		return
	}

	filename, err := app.file.Download(history)
	if err != nil {
		app.logger.Errorw("error",
			"getHistory: error downloading data to file", err,
		)
		app.errorInternalServer(w)
		return
	}

	jsonData, err := json.Marshal(filename)
	if err != nil {
		app.logger.Errorw("error",
			"getHistory: error converting filename to json", err,
		)
		app.errorInternalServer(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

type historyDownloadForm struct {
	Users []int64 `json:"user_list"`
	Days  int     `json:"days"`
}

// CreateSegment godoc
//
//	@summary        Создать сегмент
//	@description    В зависимости от параметров либо просто создает сегмент, либо создает сегмент и добавляет в него переданный процент случайно выбранных пользователей
//	@tags           segments
//	@accept         json
//	@produce        json
//	@param          slug  path    string  true    "Segment name"
//	@param          body  body    createSegmentForm  true    "Segment form"
//	@success        200 string string
//	@failure        400 {object}    errorResponse
//	@failure        500 {object}    errorResponse
//	@router         /segments/{slug} [post]
func (app *application) createSegment(w http.ResponseWriter, r *http.Request) {

	slug := chi.URLParam(r, "slug")
	ok := app.validator.SegmentSlug(slug)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	var form createSegmentForm
	err := json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		app.logger.Errorw("error",
			"createSegment: error parsing createSegmentForm", err,
		)
		app.errorWrongFormat(w)
		return
	}

	ok = app.validator.PercentageRND(form.PercentageRND)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	if err := app.storage.CreateSegment(r.Context(), slug, form.PercentageRND); err != nil {
		app.logger.Errorw("error",
			"createSegment: error inserting data to storage", err,
		)
		app.errorInternalServer(w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

type createSegmentForm struct {
	PercentageRND int `json:"percentage_random"`
}

// DeleteSegment godoc
//
//	@summary        Удалить сегмент
//	@description    Удаляет сегмент
//	@tags           segments
//	@produce        json
//	@param          slug  path    string  true    "Segment Name"
//	@success        200
//	@failure        400  {object}  errorResponse
//	@failure        500  {object}  errorResponse
//	@router         /segments/{slug} [delete]
func (app *application) deleteSegment(w http.ResponseWriter, r *http.Request) {

	slug := chi.URLParam(r, "slug")
	ok := app.validator.SegmentSlug(slug)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	if err := app.storage.DeleteSegment(r.Context(), slug); err != nil {
		app.logger.Errorw("error",
			"deleteSegment: error deleting data from storage", err,
		)
		app.errorInternalServer(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

// GetSegments godoc
//
//	@summary        Получить сегменты пользователя
//	@description    Возвращает список сегментов, в которых состоит пользователь, если таких нет, то возвращает пустой список
//	@tags           users-segments
//	@param          user_id      path    int  true    "User ID"
//	@accept         json
//	@produce        json
//	@success        200 string string
//	@failure        400 {object}    errorResponse
//	@failure        500 {object}    errorResponse
//	@router         /users-segments/{user_id} [get]
func (app *application) getSegments(w http.ResponseWriter, r *http.Request) {

	userStr := chi.URLParam(r, "user_id")

	user, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		app.errorWrongFormat(w)
		return
	}

	ok := app.validator.UserId(user)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	segments, err := app.storage.GetSegmentsByUserID(r.Context(), user)
	if err != nil {
		app.logger.Errorw("error",
			"getSegments: error retrieving data from storage", err,
		)
		app.errorInternalServer(w)
		return
	}

	jsonData, err := json.Marshal(segments)
	if err != nil {
		app.logger.Errorw("error",
			"getSegments: error converting data to json", err,
		)
		app.errorInternalServer(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}

// UpdateSegments godoc
//
//	@summary        Обновить сегменты пользователя
//	@description    Для пользователя удаляет сегменты из переданного списка, затем добавляет из второго переданного списка сегменты с указанным в днях TTL
//	@tags           users-segments
//	@accept         json
//	@produce        json
//	@param          user_id      path    int  true    "User ID"
//	@param          body    body    updateSegmentsForm    true    "Segments form"
//	@success        200 string string
//	@failure        400 {object}    errorResponse
//	@failure        500 {object}    errorResponse
//	@router         /users-segments/{user_id} [put]
func (app *application) updateSegments(w http.ResponseWriter, r *http.Request) {

	userStr := chi.URLParam(r, "user_id")
	user, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		app.errorWrongFormat(w)
		return
	}

	ok := app.validator.UserId(user)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	var form updateSegmentsForm
	err = json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		app.logger.Errorw("error",
			"updateSegments: error parsing updateSegmentsForm", err,
		)
		app.errorWrongFormat(w)
		return
	}

	ok = app.validator.Segments(form.Delete)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	ok = app.validator.Segments(form.Add)
	if !ok {
		app.errorWrongFormat(w)
		return
	}

	if err := app.storage.UpdateSegmentsByUserID(r.Context(), user, form.Delete, form.Add); err != nil {
		app.logger.Errorw("error",
			"updateSegments: error updating data in storage", err,
		)
		app.errorInternalServer(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

type updateSegmentsForm struct {
	Delete []models.Segment `json:"list_delete,omitempty"`
	Add    []models.Segment `json:"list_add,omitempty"`
}

type errorResponse struct {
	ErrorDesc Error `json:"error`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (app *application) errorNotFound(w http.ResponseWriter, r *http.Request) {
	var error errorResponse
	error.ErrorDesc.Code = http.StatusNotFound
	error.ErrorDesc.Message = "Wrong resource url"

	jsonErr, err := json.Marshal(error)
	if err != nil {
		app.logger.Errorw("error",
			"errorNotFound: error converting data to json", err,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write(jsonErr)
}

func (app *application) errorInternalServer(w http.ResponseWriter) {
	var error errorResponse
	error.ErrorDesc.Code = http.StatusInternalServerError
	error.ErrorDesc.Message = "Error while processing request. Please, contact support"

	jsonErr, err := json.Marshal(error)
	if err != nil {
		app.logger.Errorw("error",
			"errorNotFound: error converting data to json", err,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErr)
}

func (app *application) errorWrongFormat(w http.ResponseWriter) {
	var error errorResponse
	error.ErrorDesc.Code = http.StatusBadRequest
	error.ErrorDesc.Message = "Wrong body request format"

	jsonErr, err := json.Marshal(error)
	if err != nil {
		app.logger.Errorw("error",
			"errorNotFound: error converting data to json", err,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(jsonErr)
}
