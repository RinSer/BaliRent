package files

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var filesDir string

func RunFileServer() {

	filesDir = os.Getenv("FILES_DIR")

	router := mux.NewRouter()
	s := router.PathPrefix("/upload").Subrouter()
	// serve uploaded files
	_ = router.PathPrefix("/").
		Handler(http.FileServer(http.Dir(filesDir)))

	// upload files
	s.HandleFunc("", uploadFile).Methods(http.MethodPost)

	// run server
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))

}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "should upload only using POST method",
			http.StatusBadRequest)
		return
	}
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `file`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	dst, err := os.Create(filesDir + "/" + handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
