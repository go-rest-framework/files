// Package files provides ...
package files

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-rest-framework/core"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type Attachments []Attachment

type Attachment struct {
	gorm.Model
	UserID      int    `json:"userID"`
	Group       string `json:"group"`
	FileID      int    `json:"fileID"`
	Title       string `json:"title"`
	Description string `json:"description" gorm:"type:text"`
	IsMain      int    `json:"isMain"`
	Hash        string `json:"hash"`
	Index       int    `json:"index" gorm:"type:int(6)"`
	File        File   `json:"file"`
}

func actionAttchGetAll(w http.ResponseWriter, r *http.Request) {
	var (
		attachments Attachments
		rsp         = core.Response{Data: &attachments}
		all         = r.FormValue("all")
		id          = r.FormValue("id")
		group       = r.FormValue("group")
		fileid      = r.FormValue("fileid")
		title       = r.FormValue("title")
		description = r.FormValue("description")
		hash        = r.FormValue("hash")
		sort        = r.FormValue("sort")
		limit       = r.FormValue("limit")
		offset      = r.FormValue("offset")
		db          = App.DB
	)

	if all != "" {
		db = db.Where("id LIKE ?", "%"+all+"%")
		db = db.Or("title LIKE ?", "%"+all+"%")
		db = db.Or("description LIKE ?", "%"+all+"%")
		db = db.Or("content LIKE ?", "%"+all+"%")
	}

	if id != "" {
		db = db.Where("id = ?", id)
	}

	if group != "" {
		db = db.Where("`group` = ?", group)
	}

	if fileid != "" {
		db = db.Where("file_id = ?", fileid)
	}

	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
	}

	if description != "" {
		db = db.Where("description LIKE ?", "%"+description+"%")
	}

	if hash != "" {
		db = db.Where("hash = ?", hash)
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

	db.Find(&attachments)

	rsp.Data = &attachments

	w.Write(rsp.Make())
}

func actionAttachGetOne(w http.ResponseWriter, r *http.Request) {
	var (
		attachment Attachment
		rsp        = core.Response{Data: &attachment}
		db         = App.DB
	)

	vars := mux.Vars(r)

	db = db.Set("gorm:auto_preload", true)
	db = db.Preload("File")

	db.First(&attachment, vars["id"])

	if attachment.ID == 0 {
		rsp.Errors.Add("ID", "Attachment not found")
	} else {
		rsp.Data = &attachment
	}

	w.Write(rsp.Make())
}

func actionAttachCreate(w http.ResponseWriter, r *http.Request) {
	var (
		attachment Attachment
		rsp        = core.Response{Data: &attachment}
	)

	if rsp.IsJsonParseDone(r.Body) {
		if rsp.IsValidate() {
			userid, _ := strconv.Atoi(r.Header.Get("id"))
			attachment.UserID = userid
			App.DB.Create(&attachment)
		}
	}

	rsp.Data = &attachment

	w.Write(rsp.Make())
}

func actionAttachUpdate(w http.ResponseWriter, r *http.Request) {
	var (
		data       Attachment
		attachment Attachment
		rsp        = core.Response{Data: &data}
	)

	if rsp.IsJsonParseDone(r.Body) {
		if rsp.IsValidate() {

			vars := mux.Vars(r)
			App.DB.First(&attachment, vars["id"])

			if attachment.ID == 0 {
				rsp.Errors.Add("ID", "Attachment not found")
			} else {
				role := r.Header.Get("role")
				idstring := fmt.Sprintf("%d", attachment.UserID)
				userid := r.Header.Get("id")
				if role == "admin" || (role == "user" && idstring == userid) {
					App.DB.Model(&attachment).Updates(data)
				} else {
					rsp.Errors.Add("ID", "Only owner can change attachment")
				}
			}
		}
	}

	rsp.Data = &attachment

	w.Write(rsp.Make())
}

func actionAttachDelete(w http.ResponseWriter, r *http.Request) {
	var (
		attachment Attachment
		rsp        = core.Response{Data: &attachment}
	)

	vars := mux.Vars(r)
	App.DB.First(&attachment, vars["id"])

	if attachment.ID == 0 {
		rsp.Errors.Add("ID", "Contentattachment not found")
	} else {
		role := r.Header.Get("role")
		idstring := fmt.Sprintf("%d", attachment.UserID)
		userid := r.Header.Get("id")
		if role == "admin" || (role == "user" && idstring == userid) {
			if App.IsTest {
				App.DB.Unscoped().Delete(&attachment)
			} else {
				App.DB.Delete(&attachment)
			}
		} else {
			rsp.Errors.Add("ID", "Only owner can delete attachment")
		}
	}

	rsp.Data = &attachment

	w.Write(rsp.Make())
}
