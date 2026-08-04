package main

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type HttpConfig struct {
    Url     string
    Port    string
}

type DbConfig struct {
    Url     string
    Name    string
}

type Nh struct {
    API         string
    SyncMinutes int
    Enabled     bool
}

type ConfigT struct {
    Http        HttpConfig
    Dbase       DbConfig
    Nh          Nh
}

var Config = ConfigT{
    Http: HttpConfig{
        Url:    "",
        Port:   "3002",
    },
    Dbase: DbConfig{
        Url:    "mongodb://localhost:27017",
        Name:   "nihility",
    },
    Nh: Nh{
        API: "",
        SyncMinutes: 1,
        Enabled: false,
    },
}

func InitConfig() {
    ex, err := os.Executable()
    if nil != err {
        panic(err)
    }
    expath := filepath.Dir(ex)
    configfile := expath + "/.config.json"

    dat, err := os.ReadFile(configfile)
    if nil != err {
        configdat, _ := json.MarshalIndent(Config, "", "  ")
        os.WriteFile(configfile, configdat, 0644)
    } else {
        json.Unmarshal(dat, &Config)
    }
}
