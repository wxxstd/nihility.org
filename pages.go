package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/asapgiri/golib/renderer"
	"github.com/asapgiri/golib/session"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetCurrentSession(w http.ResponseWriter, r *http.Request) session.Sessioner {
    sess := session.Sessioner{}
    sess.Authenticate(w, r)
    Authenticate(&sess.Auth)

    sess.Path = r.URL.Path

    sess.Meta = session.MetaData{}

    return sess
}

func Login(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" != session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    uname := r.FormValue("form[userName]")
    upass := r.FormValue("form[userPass]")

    if "" != uname {
        user := Staff{}
        err := user.Login(uname, upass)
        log.Println(user)
        log.Println(err)
        if nil != err {
            session.SetError(err.Error())
        } else {
            session.Delete(w, r)
            session.New(w, r, user.Nick)
        }
    } else {
        session.SetError("")
    }

    if "" == session.Auth.Username {
        fil, _ := renderer.ReadArtifact("login.html", w.Header())
        renderer.Render(session, w, fil, nil)
    } else {
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func Register(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" != session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sha := RegisterSha{}
	sha.Find(r.PathValue("sha"))
	if "" == sha.Sha {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    uuname := r.FormValue("form[userUsername]")
    upassa := r.FormValue("form[userPassA]")
    upassb := r.FormValue("form[userPassB]")

    if "" != uuname {
        user := Staff{
            Nick: uuname,
        }

        oid, err := primitive.ObjectIDFromHex(sha.Sha)
        if nil == err {
            user.Select(oid)
        }

        err = user.Register(upassa, upassb)
        if nil != err {
            session.SetError(err.Error())
        } else {
            session.Delete(w, r)
            session.New(w, r, user.Nick)
	        sha.Delete()
        }
    } else {
        session.SetError("")
    }

    if "" == session.Auth.Username {
        fil, _ := renderer.ReadArtifact("register.html", w.Header())
        renderer.Render(session, w, fil, sha)
    } else {
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func Logout(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)
    session.Delete(w, r)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Works(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    list := collect_translations()

    fil, _ := renderer.ReadArtifact("works.html", w.Header())
    renderer.Render(session, w, fil, list)
}

func WorkAdd(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    tr := Translation{
        From:   r.FormValue("form[from]"),
        Title:  r.FormValue("form[title]"),
        NhId:   r.FormValue("form[nhid]"),
        EhId:   r.FormValue("form[ehid]"),
        Date:   time.Now(),
    }

    if "" != tr.Title || "" != tr.NhId {
        if "" != tr.NhId {
            inf := get_info(tr.NhId)

            // TODO: Should I add more?
            tr.Date = time.Unix(inf.UploadDate, 0)
            if "" == tr.Title {
                tr.Title = inf.Title.English
            }
            if "" == tr.From {
                tr.From = inf.Title.Japanese
            }
        }

        tr.Id = primitive.NewObjectID()
        tr.Add()
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
    }


    fil, _ := renderer.ReadArtifact("workadd.html", w.Header())
    renderer.Render(session, w, fil, nil)
}

func WorkEdit(w http.ResponseWriter, r *http.Request) {
	session := GetCurrentSession(w, r)

	if session.Auth.Username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.PathValue("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	tr := Translation{
		Id: objID,
	}

	// GET existing data
	if r.Method == http.MethodGet {
		err := tr.Select(objID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

        tr.sync_info()

		fil, _ := renderer.ReadArtifact("workedit.html", w.Header())
		renderer.Render(session, w, fil, tr)
		return
	}

	// POST update
	if r.Method == http.MethodPost {

		err := tr.Select(objID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		tr.From = r.FormValue("form[from]")
		tr.Title = r.FormValue("form[title]")
		tr.Banner = r.FormValue("form[banner]")
		tr.NhId = r.FormValue("form[nhid]")
		tr.EhId = r.FormValue("form[ehid]")

		date, err := time.Parse(
			"2006-01-02T15:04",
			r.FormValue("form[date]"),
		)

		if err == nil {
			tr.Date = date
		}

		tr.Update()

		http.Redirect(w, r, "/works", http.StatusSeeOther)
		return
	}
}

func WorkDelete(w http.ResponseWriter, r *http.Request) {
	session := GetCurrentSession(w, r)

	if session.Auth.Username == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.PathValue("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	tr := Translation{
		Id: objID,
	}

    err = tr.Select(objID)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    tr.Delete()
	http.Redirect(w, r, "/works", http.StatusSeeOther)
}

func StaffList(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    var staff Staff
    list, _ := staff.List()

    fil, _ := renderer.ReadArtifact("staff.html", w.Header())
    renderer.Render(session, w, fil, list)
}

func StaffInvite(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.PathValue("id")

	sha := RegisterSha{}
	sha.Find(id)
	if "" == sha.Sha {
        sha.Id = primitive.NewObjectID()
        sha.Sha = id

        var staff Staff
	    objID, err := primitive.ObjectIDFromHex(id)
        if nil == err && nil == staff.Select(objID) {
            sha.Staff = staff
        }

        sha.Add()
	}

    fil, _ := renderer.ReadArtifact("invite.html", w.Header())
    renderer.Render(session, w, fil, sha)
}

func get_artifact_images() []string {
    files, _ := os.ReadDir("artifacts")

    result := []string{}

    for _, file := range files {
        if file.IsDir() {
            continue
        }

        result = append(result, "/"+file.Name())
    }

    return result
}

func StaffSite(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    if r.Method == http.MethodPost {

        bannerPath := ""

        file, header, err := r.FormFile("banner")
        if err == nil {
            defer file.Close()

            ext := filepath.Ext(header.Filename)

            filename := primitive.NewObjectID().Hex() + ext

            path := "artifacts/" + filename

            dst, err := os.Create(path)
            if err == nil {
                defer dst.Close()

                io.Copy(dst, file)

                bannerPath = "/" + filename
            }
        }


        update_setting("title", r.FormValue("form[title]"))
        update_setting("discord", r.FormValue("form[discord]"))

        if bannerPath != "" {
            update_setting("banner", bannerPath)
        } else {
            update_setting("banner", r.FormValue("form[banner]"))
        }

        update_setting("intro", r.FormValue("form[intro]"))

        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    type dto_site struct {
        Site        settings
        Artifacts   []string
    }
    dto := dto_site{
        Site: get_site_settings(),
        Artifacts: get_artifact_images(),
    }

    fil, _ := renderer.ReadArtifact("site.html", w.Header())
    renderer.Render(session, w, fil, dto)
}
