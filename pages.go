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

    user := Staff{
        Nick: uuname,
    }

    if !sha.Staff.IsZero() {
        user.Select(sha.Staff)
    }

    if "" != uuname {
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
        renderer.Render(session, w, fil, user)
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

        selected := make(map[string]bool)
        for _, s := range tr.Staff {
            selected[s.Nick] = true
            for _, rol := range s.Role {
                selected[s.Nick+rol.Id.Hex()] = true
            }
        }

        var staff Staff
        var role Role

        type dto_tr struct {
            Translation Translation
            Staff       []Staff
            Roles       []Role
            Selected    map[string]bool
        }
        dto := dto_tr{
            Translation: tr,
            Selected: selected,
        }
        dto.Staff, _ = staff.List()
        dto.Roles, _ = role.List()

		fil, _ := renderer.ReadArtifact("workedit.html", w.Header())
		renderer.Render(session, w, fil, dto)
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

        var role Role
        roles, _ := role.List()
        form_staff := r.Form["form[staff]"]
        tr.Staff = make([]TrStaff, len(form_staff))
        for i, nick := range form_staff {
            tr.Staff[i] = TrStaff{
                Nick: nick,
                Role: []Role{},
            }
            form_roles := r.Form["form[roles]["+nick+"]"]
            for _, frl := range form_roles {
                for _, rl := range roles {
                    if frl == rl.Id.Hex() {
                        tr.Staff[i].Role = append(tr.Staff[i].Role, rl)
                    }
                }
            }
        }


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

    var staff Staff
	sha := RegisterSha{}
	sha.Find(id)

    objID, _ := primitive.ObjectIDFromHex(id)
    staff.Select(objID)


	if "" == sha.Sha {
        sha.Id = primitive.NewObjectID()
        sha.Sha = id
        sha.Staff = staff.Id
        sha.Add()
	}

    type dto_inv struct {
        Sha     RegisterSha
        Staff   Staff
    }
    dto := dto_inv{
        Sha: sha,
        Staff: staff,
    }

    fil, _ := renderer.ReadArtifact("invite.html", w.Header())
    renderer.Render(session, w, fil, dto)
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

func save_file_form(r *http.Request, from string) string {
    file, header, err := r.FormFile(from)

    if err == nil {
        defer file.Close()

        ext := filepath.Ext(header.Filename)
        filename := primitive.NewObjectID().Hex() + ext
        path := "artifacts/" + filename

        dst, err := os.Create(path)
        if err == nil {
            defer dst.Close()
            io.Copy(dst, file)

            return "/" + filename
        }
    }

    return ""
}

func StaffSite(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    if r.Method == http.MethodPost {

        bannerPath := save_file_form(r, "banner")

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

func StaffAdd(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    if r.Method == http.MethodPost {

        avatarPath := save_file_form(r, "avatar")

        var avatar string
        if "" != avatarPath {
            avatar = avatarPath
        } else {
            avatar = r.FormValue("form[avatar]")
        }

        staff := Staff{
            Avatar:  avatar,
            Nick:    r.FormValue("form[nick]"),
            Discord: r.FormValue("form[discord]"),
            Email:   r.FormValue("form[email]"),
        }
        staff.Id = primitive.NewObjectID()
        staff.Add()

        http.Redirect(w, r, "/staff", http.StatusSeeOther)
        return
    }

    type dto_sa struct {
        Artifacts   []string
    }
    dto := dto_sa{
        Artifacts: get_artifact_images(),
    }

    fil, _ := renderer.ReadArtifact("staffadd.html", w.Header())
    renderer.Render(session, w, fil, dto)
}

func StaffEdit(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

	if "" == session.Auth.Username {
        http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    oid, _ := primitive.ObjectIDFromHex(r.PathValue("id"))
    staff := Staff{}
    staff.Select(oid)

    if r.Method == http.MethodPost {

        avatarPath := save_file_form(r, "avatar")

        var avatar string
        if "" != avatarPath {
            avatar = avatarPath
        } else {
            avatar = r.FormValue("form[avatar]")
        }

        staff.Avatar =  avatar
        staff.Nick =    r.FormValue("form[nick]")
        staff.Discord = r.FormValue("form[discord]")
        staff.Email =   r.FormValue("form[email]")
        staff.Update()

        http.Redirect(w, r, "/staff", http.StatusSeeOther)
        return
    }

    type dto_sa struct {
        Artifacts   []string
        Staff       Staff
    }
    dto := dto_sa{
        Artifacts: get_artifact_images(),
        Staff: staff,
    }

    fil, _ := renderer.ReadArtifact("staffedit.html", w.Header())
    renderer.Render(session, w, fil, dto)
}

func StaffDelete(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

    if session.Auth.Username == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    oid, _ := primitive.ObjectIDFromHex(r.PathValue("id"))
    staff := Staff{}
    staff.Select(oid)
    staff.Delete()

    http.Redirect(w, r, "/staff", http.StatusSeeOther)
}

func Roles(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

    if session.Auth.Username == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var role Role
    roles, _ := role.List()

    fil, _ := renderer.ReadArtifact("roles.html", w.Header())
    renderer.Render(session, w, fil, roles)
}

func RolesAdd(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

    if session.Auth.Username == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    if r.Method == http.MethodPost {
        role := Role{
            Id:    primitive.NewObjectID(),
            Name:  r.FormValue("form[name]"),
            Color: r.FormValue("form[color]"),
            Desc:  r.FormValue("form[desc]"),
        }

        if role.Name != "" {
            role.Add()
        }

        http.Redirect(w, r, "/roles", http.StatusSeeOther)
        return
    }

    fil, _ := renderer.ReadArtifact("rolesedit.html", w.Header())
    renderer.Render(session, w, fil, Role{})
}

func RolesEdit(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

    if session.Auth.Username == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
    if err != nil {
        http.Redirect(w, r, "/roles", http.StatusSeeOther)
        return
    }

    var role Role
    if role.Select(id) != nil {
        http.Redirect(w, r, "/roles", http.StatusSeeOther)
        return
    }

    if r.Method == http.MethodPost {
        role.Name = r.FormValue("form[name]")
        role.Color = r.FormValue("form[color]")
        role.Desc = r.FormValue("form[desc]")
        role.Update()

        http.Redirect(w, r, "/roles", http.StatusSeeOther)
        return
    }

    fil, _ := renderer.ReadArtifact("rolesedit.html", w.Header())
    renderer.Render(session, w, fil, role)
}

func RolesDelete(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

    if session.Auth.Username == "" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
    if err != nil {
        http.Redirect(w, r, "/roles", http.StatusSeeOther)
        return
    }

    var role Role
    if role.Select(id) == nil {
        role.Delete()
    }

    http.Redirect(w, r, "/roles", http.StatusSeeOther)
}

