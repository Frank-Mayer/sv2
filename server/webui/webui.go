package webui

import (
	"net/http"
	"os"

	"github.com/charmbracelet/log"
)

const (
	documentPath = "/binds/dashboard.html"
)

var (
	// document is the HTML document for the web UI.
	document string
)

func loadDocument() error {
	if document != "" {
		return nil
	}
	if fileContent, err := os.ReadFile(documentPath); err != nil {
		return err
	} else {
		document = string(fileContent)
	}
	return nil
}

func Dashboard(w http.ResponseWriter, _ *http.Request) {
	if err := loadDocument(); err != nil {
		log.Error("Failed to load document", "path", documentPath, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(document)); err != nil {
		log.Error("Failed to write document", "error", err)
	}
}
