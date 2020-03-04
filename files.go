package files

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	Src    string `json:"src"`
	Ext    string `json:"ext" gorm:"type:varchar(10)"`
	Preset string `json:"preset"`
	Size   int64  `json:"size"`
	Status int    `json:"status"`
	Type   int    `json:"type"`
	Hash   string `json:"hash"`
}

func Configure(a core.App) {
	App = a

	App.DB.AutoMigrate(&File{}, &Attachment{})

	//public actions

	//protect CRUD actions with files info
	App.R.HandleFunc("/files", actionGetAll).Methods("GET")
	App.R.HandleFunc("/files/{id}", actionGetOne).Methods("GET")
	App.R.HandleFunc(
		"/files",
		App.Protect(
			actionUpload,
			[]string{"admin", "user"})).Methods("POST")
	App.R.HandleFunc(
		"/files/{id}",
		App.Protect(
			actionReUpload,
			[]string{"admin", "user"})).Methods("PATCH")
	App.R.HandleFunc(
		"/files/{id}",
		App.Protect(
			actionDelete,
			[]string{"admin", "user"})).Methods("DELETE")

	App.R.HandleFunc("/attachments", actionAttchGetAll).Methods("GET")
	App.R.HandleFunc("/attachments/{id}", actionAttachGetOne).Methods("GET")
	App.R.HandleFunc(
		"/attachments",
		App.Protect(
			actionAttachCreate,
			[]string{"admin", "user"})).Methods("POST")
	App.R.HandleFunc(
		"/attachments/{id}",
		App.Protect(
			actionAttachUpdate,
			[]string{"admin", "user"})).Methods("PATCH")
	App.R.HandleFunc(
		"/attachments/{id}",
		App.Protect(
			actionAttachDelete,
			[]string{"admin", "user"})).Methods("DELETE")
}

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
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

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return File{}, err
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	tm := time.Now()
	d := 24 * time.Hour

	userpersonaldir := fmt.Sprintf("%x", md5.Sum([]byte(r.Header.Get("id"))))
	userdatedir := fmt.Sprintf("%d", tm.Truncate(d).Unix())

	userdir := dir + "/" + App.Config.WebRootPath + "/" + App.Config.UploadsPath
	userdir += "/" + userpersonaldir
	userdir += "/" + userdatedir

	CreateDirIfNotExist(userdir)

	err = ioutil.WriteFile(userdir+"/"+filename+fileext, fileBytes, 0644)

	if err != nil {
		return File{}, err
	}

	userid, _ := strconv.Atoi(r.Header.Get("id"))

	return File{
		UserID: userid,
		Name:   filename,
		Path:   userdir + "/" + filename + fileext,
		Src:    "/" + App.Config.UploadsPath + "/" + userpersonaldir + "/" + userdatedir + "/" + filename + fileext,
		Ext:    fileext,
		Preset: "notset",
		Size:   handler.Size,
		Status: 0,
		Type:   0,
		Hash:   fmt.Sprintf("%x", md5.Sum(fileBytes)),
	}, nil
}

func actionGetAll(w http.ResponseWriter, r *http.Request) {
	var (
		files  Files
		rsp    = core.Response{Data: &files, Req: r}
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
		db = db.Where("id = ?", id)
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
		rsp  = core.Response{Data: &file, Req: r}
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
		rsp       = core.Response{Data: &filemodel, Req: r}
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
		rsp       = core.Response{Data: &filemodel, Req: r}
	)

	vars := mux.Vars(r)
	App.DB.First(&filemodel, vars["id"])

	if filemodel.ID == 0 {
		rsp.Errors.Add("ID", "File not found")
	} else {
		role := r.Header.Get("role")
		idstring := fmt.Sprintf("%d", filemodel.UserID)
		userid := r.Header.Get("id")
		if role == "admin" || (role == "user" && idstring == userid) {
			data, err := upload(r)
			if err != nil {
				rsp.Errors.Add("file", err.Error())
			} else {
				err := os.Remove(filemodel.Path)
				if err != nil {
					rsp.Errors.Add("file", err.Error())
				}
				App.DB.Model(&filemodel).Updates(data)
			}
		} else {
			rsp.Errors.Add("file", "Only owner can change element")
		}
	}

	w.Write(rsp.Make())
}

func actionDelete(w http.ResponseWriter, r *http.Request) {
	var (
		file File
		rsp  = core.Response{Data: &file, Req: r}
	)

	vars := mux.Vars(r)
	App.DB.First(&file, vars["id"])

	if file.ID == 0 {
		rsp.Errors.Add("ID", "File not found")
	} else {
		role := r.Header.Get("role")
		idstring := fmt.Sprintf("%d", file.UserID)
		userid := r.Header.Get("id")
		if role == "admin" || (role == "user" && idstring == userid) {
			if App.IsTest {
				App.DB.Unscoped().Delete(&file)
			} else {
				App.DB.Delete(&file)
			}
			err := os.Remove(file.Path)
			if err != nil {
				rsp.Errors.Add("file", err.Error())
			}
		} else {
			rsp.Errors.Add("file", "Only owner can delete element")
		}
	}

	rsp.Data = &file

	w.Write(rsp.Make())
}
