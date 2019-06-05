package files

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-rest-framework/core"
)

var App core.App

func Configure(a core.App) {
	App = a

	//App.DB.AutoMigrate(&User{}, &Profile{})

	//public actions

	//protect CRUD actions with files info
	//App.R.HandleFunc("/api/files", App.Protect(actionGetAll, []string{"admin"})).Methods("GET")
	//App.R.HandleFunc("/api/files/{id}", App.Protect(actionGetOne, []string{"admin"})).Methods("GET")
	App.R.HandleFunc("/api/files", actionUpload).Methods("POST")
	//App.R.HandleFunc("/api/files/{id}", App.Protect(actionUpdate, []string{"admin"})).Methods("PATCH")
	//App.R.HandleFunc("/api/files/{id}", App.Protect(actionDelete, []string{"admin"})).Methods("DELETE")
}

func actionUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("web/uploads", "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}
