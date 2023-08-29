package file

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/h3ll0kitt1/avitotest/internal/models"
)

type File interface {
	Download(history []models.History) (string, error)
}

type FileCSV struct {
	filename string
}

func NewCSV(filename string) *FileCSV {
	return &FileCSV{filename: filename}
}

func (f *FileCSV) Download(history []models.History) (string, error) {

	csvFile, err := os.Create(f.filename)
	if err != nil {
		return "", err
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	for _, record := range history {
		var row []string

		user := strconv.FormatInt(record.User, 10)
		row = append(row, user)
		row = append(row, record.Segment.Slug)

		if record.Action {
			row = append(row, "добавление")
		}

		if !record.Action {
			row = append(row, "удаление")
		}

		row = append(row, record.ActionTime)
		writer.Write(row)
	}
	writer.Flush()

	return f.filename, nil
}
