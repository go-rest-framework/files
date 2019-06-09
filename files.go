package files

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/go-rest-framework/core"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var App core.App

type Files []File

type File struct {
	gorm.Model
	UserID int    `json:"userID"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Ext    string `json:"ext" gorm:"type:varchar(10)"`
	Preset string `json:"preset"`
	Size   int64  `json:"size"`
	Status int    `json:"status"`
	Type   int    `json:"type"`
	Hash   string `json:"hash"`
}

type Attachment struct {
	gorm.Model
	Group       string `json:"group"`
	FileID      int    `json:"fileID"`
	Title       string `json:"title"`
	Description string `json:"description" gorm:"type:text"`
	IsMain      int    `json:"isMain"`
	Bind        string `json:"bind"`
	Hash        string `json:"hash"`
	Index       int    `json:"index" gorm:"type:int(6)"`
}

func Configure(a core.App) {
	App = a

	App.DB.AutoMigrate(&File{}, &Attachment{})

	//public actions

	//protect CRUD actions with files info
	App.R.HandleFunc("/api/files", actionGetAll).Methods("GET")
	App.R.HandleFunc("/api/files/{id}", actionGetOne).Methods("GET")
	App.R.HandleFunc("/api/files", actionUpload).Methods("POST")
	App.R.HandleFunc("/api/files/{id}", actionReUpload).Methods("PATCH")
	App.R.HandleFunc("/api/files/{id}", actionDelete).Methods("DELETE")
}

func actionGetAll(w http.ResponseWriter, r *http.Request) {
	var (
		files  Files
		rsp    = core.Response{Data: &files}
		all    = r.FormValue("all")
		id     = r.FormValue("id")
		name   = r.FormValue("name")
		path   = r.FormValue("path")
		ext    = r.FormValue("ext")
		preset = r.FormValue("preset")
		sort   = r.FormValue("sort")
		limit  = r.FormValue("limit")
		offset = r.FormValue("offset")
		db     = App.DB
	)

	if all != "" {
		db = db.Where("id LIKE ?", "%"+all+"%")
		db = db.Or("name LIKE ?", "%"+all+"%")
		db = db.Or("path LIKE ?", "%"+all+"%")
		db = db.Or("ext LIKE ?", "%"+all+"%")
	}

	if id != "" {
		db = db.Where("id LIKE ?", "%"+id+"%")
	}

	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}

	if path != "" {
		db = db.Where("path LIKE ?", "%"+path+"%")
	}

	if ext != "" {
		db = db.Where("ext LIKE ?", "%"+ext+"%")
	}

	if preset != "" {
		db = db.Where("preset = ?", preset)
	}

	if sort != "" {
		db = db.Order(sort)
	}

	if limit != "" {
		db = db.Limit(limit)
	}

	if offset != "" {
		db = db.Offset(offset)
	}

	db.Find(&files)

	rsp.Data = &files

	w.Write(rsp.Make())
}

func actionGetOne(w http.ResponseWriter, r *http.Request) {
	var (
		file File
		rsp  = core.Response{Data: &file}
		db   = App.DB
	)

	vars := mux.Vars(r)

	db.First(&file, vars["id"])

	if file.ID == 0 {
		rsp.Errors.Add("ID", "File not found")
	} else {
		rsp.Data = &file
	}

	w.Write(rsp.Make())
}

func actionUpload(w http.ResponseWriter, r *http.Request) {
	var (
		filemodel File
		rsp       = core.Response{Data: &filemodel}
	)

	filemodel, err := upload(r)
	if err != nil {
		rsp.Errors.Add("file", err.Error())
	} else {
		App.DB.Create(&filemodel)

		rsp.Data = &filemodel
	}

	w.Write(rsp.Make())
}

func actionReUpload(w http.ResponseWriter, r *http.Request) {
	var (
		filemodel File
		rsp       = core.Response{Data: &filemodel}
	)

	data, err := upload(r)
	if err != nil {
		rsp.Errors.Add("file", err.Error())
	} else {
		vars := mux.Vars(r)
		App.DB.First(&filemodel, vars["id"])

		if filemodel.ID == 0 {
			rsp.Errors.Add("ID", "Contentelement not found")
		} else {
			//idstring := fmt.Sprintf("%d", filemodel.UserID)
			//if idstring != r.Header.Get("id") {
			//rsp.Errors.Add("ID", "Only owner can change element")
			//} else {
			App.DB.Model(&filemodel).Updates(data)
			//}
		}
	}

	w.Write(rsp.Make())
}

func actionDelete(w http.ResponseWriter, r *http.Request) {
	var (
		file File
		rsp  = core.Response{Data: &file}
	)

	vars := mux.Vars(r)
	App.DB.First(&file, vars["id"])

	if file.ID == 0 {
		rsp.Errors.Add("ID", "File not found")
	} else {
		if App.IsTest {
			App.DB.Unscoped().Delete(&file)
		} else {
			App.DB.Delete(&file)
		}
	}

	rsp.Data = &file

	w.Write(rsp.Make())
}

func upload(r *http.Request) (File, error) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		return File{}, errors.New("Error Retrieving the File")
	}
	defer file.Close()
	filename := strings.TrimSuffix(handler.Filename, path.Ext(handler.Filename))
	fileext := path.Ext(handler.Filename)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("web/uploads", "*"+filename+fileext)
	if err != nil {
		return File{}, err
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return File{}, err
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	return File{
		UserID: 0,
		Name:   filename,
		Path:   tempFile.Name(),
		Ext:    fileext,
		Preset: "notset",
		Size:   handler.Size,
		Status: 0,
		Type:   0,
		Hash:   fmt.Sprintf("%x", md5.Sum(fileBytes)),
	}, nil
}
