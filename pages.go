package main

import (
	"net/http"
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
        err := user.Register(upassa, upassb)
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
        renderer.Render(session, w, fil, nil)
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

    if "" != tr.Title {
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

    if "" != tr.Title {
        tr.Id = primitive.NewObjectID()
        tr.Add()
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
    }


    fil, _ := renderer.ReadArtifact("workadd.html", w.Header())
    renderer.Render(session, w, fil, nil)
}
