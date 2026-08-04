package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/asapgiri/golib/logger"
	"github.com/asapgiri/golib/renderer"
	"github.com/asapgiri/golib/session"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var log = logger.Logger {
    Color: logger.Colors.Purple,
    Pretext: "nihility",
}

type RegisterSha struct {
    Id      primitive.ObjectID `bson:"_id"`
	Sha     string
    Staff   Staff
}

type Setting struct {
    Id      primitive.ObjectID `bson:"_id"`
    Name    string
    Value   string
}

type Role struct {
    Id      primitive.ObjectID `bson:"_id"`
    Name    string
    Color   string
}

type Staff struct {
    Id      primitive.ObjectID `bson:"_id"`
    IsUser  bool
    Avatar  string
    Nick    string
    Discord string
    Email   string
    Passwd  string
    Roles   []Role
}

type Translation struct {
    Id      primitive.ObjectID `bson:"_id"`
    From    string
    Title   string
    Banner  string
    NhId    string
    NhLikes int
    EhId    string
    Date    time.Time
    Staff   []Staff
}

type settings struct {
    Title   string
    Discord string
    Intro   string
    Banner  string
}

type dto struct {
    Settings        settings
    Staff           []Staff
    Translations    []Translation
}

type NhentaiGallery struct {
	ID           int    `json:"id"`
	MediaID      string `json:"media_id"`
	UploadDate   int64  `json:"upload_date"`
	NumPages     int    `json:"num_pages"`
	NumFavorites int    `json:"num_favorites"`

	Title struct {
		English  string `json:"english"`
		Japanese string `json:"japanese"`
		Pretty   string `json:"pretty"`
	} `json:"title"`

	Cover struct {
		Path string `json:"path"`
	} `json:"cover"`

	Tags []struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"tags"`
}

func Unexpected(sess session.Sessioner, w http.ResponseWriter, r *http.Request) {
    fil, typ := renderer.ReadArtifact(r.URL.Path, w.Header())

    if "text" == typ {
        renderer.Render(sess, w, fil, nil)
    } else {
        io.WriteString(w, fil)
    }
}

type info struct {
    LastUpdate  time.Time
    Value       NhentaiGallery
}
var info_cache = make(map[string]info)

func get_info(id string) NhentaiGallery {
    var gallery NhentaiGallery

    val, ok := info_cache[id]
    if ok && val.LastUpdate.Sub(time.Now()).Minutes() < float64(Config.Nh.SyncMinutes) {
        return val.Value
    }

    resp, err := http.Get("https://nhentai.net/api/v2/galleries/"+id)
    if err != nil {
        log.Println(err)
        return gallery
    }
    defer resp.Body.Close()

    err = json.NewDecoder(resp.Body).Decode(&gallery)
    if err != nil {
        log.Println(err)
    }

    info_cache[id] = info{
        LastUpdate: time.Now(),
        Value: gallery,
    }

    return gallery
}

func (tr *Translation) sync_info() {
    inf := get_info(tr.NhId)
    tr.NhLikes = inf.NumFavorites
    tr.Banner = inf.Cover.Path
}

func collect_translations() []Translation {
    var tr Translation
    translations, _ := tr.List()

    for i, _ := range translations {
        translations[i].sync_info()
    }

    sort.Slice(translations, func(i, j int) bool {
        return translations[i].Date.After(translations[j].Date)
    })

    return translations
}

func get_site_settings() settings {
    result := settings{}
    stt := Setting{}

    // Get all settings from DB
    list, err := stt.List()
    if err != nil {
        return result
    }

    for _, stt := range list {
        switch stt.Name {
        case "title":
            result.Title = stt.Value
        case "discord":
            result.Discord = stt.Value
        case "banner":
            result.Banner = stt.Value
        case "intro":
            result.Intro = stt.Value
        }

    }

    return result
}

func update_setting(name, value string) {
    var s Setting
    s.FindByName(name)
    s.Name = name
    s.Value = value
    s.Update()
}

func Root(w http.ResponseWriter, r *http.Request) {
    session := GetCurrentSession(w, r)

    if "/" == r.URL.Path {
        var staff Staff
        staff_list, _ := staff.List()

        dto_tr := dto{
            Staff: staff_list,
            Translations: collect_translations(),
            Settings: get_site_settings(),
        }

        fil, _ := renderer.ReadArtifact("index.html", w.Header())
        renderer.Render(session, w, fil, dto_tr)
    } else {
        Unexpected(session, w, r)
    }
}

func main() {
    dbConnect()
    InitConfig()

    http.HandleFunc("GET /",                    Root)
    http.HandleFunc("GET /index",               Root)
    http.HandleFunc("GET /index.html",          Root)

    http.HandleFunc("GET /login",               Login)
    http.HandleFunc("POST /login",              Login)
    http.HandleFunc("GET /register/{sha}",      Register)
    http.HandleFunc("POST /register/{sha}",     Register)
    http.HandleFunc("GET /logout",              Logout)

    http.HandleFunc("GET /works",               Works)
    http.HandleFunc("GET /works/add",           WorkAdd)
    http.HandleFunc("POST /works/add",          WorkAdd)
    http.HandleFunc("GET /works/edit/{id}",     WorkEdit)
    http.HandleFunc("POST /works/edit/{id}",    WorkEdit)
    http.HandleFunc("GET /works/delete/{id}",   WorkDelete)

    http.HandleFunc("GET /staff",               StaffList)
    http.HandleFunc("GET /staff/invite/{id}",   StaffInvite)
    http.HandleFunc("GET /staff/site",          StaffSite)
    http.HandleFunc("POST /staff/site",         StaffSite)

    args := os.Args[1:]
    if 0 < len(args) {
        Config.Http.Port = args[0];
    }

    http.ListenAndServe(strings.Join([]string{":", Config.Http.Port}, ""), nil)
}
