package mashapp

import (
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "net/http"    
    "strings"
)

type MashAppServer struct {
    // final
    port int
    rootPath string
    filePath string
    state *ServerState

    // Lazily loaded
    fileOptions []string
}

func NewServer(port int, rootPath string, filePath string) *MashAppServer {
    return &MashAppServer{
        port, 
        rootPath,
        filePath,
        NewServerState(),
        nil,
    }
}

func (s *MashAppServer) Serve() {
    addr := fmt.Sprintf(":%d", s.port)
    fmt.Printf("Serving http://localhost%s/\n", addr)

    serveStaticFiles(fmt.Sprintf("%s/static", s.rootPath), "static")
    serveStaticFiles(fmt.Sprintf("%s/polymer/app", s.rootPath), "polymer")

    s.serveRPCs()

    http.HandleFunc("/", s.appHandler)
    
    http.ListenAndServe(addr, nil)
}

type BootstrapData struct {
    Files []string `json:"files"`
}
type TemplateData struct {
    // Files []string
    JsonConfig template.HTML
}

func (s *MashAppServer) appHandler(w http.ResponseWriter, r *http.Request) {
    if s.fileOptions == nil {
        s.fileOptions = listMusicFiles(s.filePath)
    }

    bootstrap := BootstrapData {
        s.fileOptions,
    }
    asJson, _ := json.Marshal(bootstrap)
    s.renderTemplate(w, "app.html", TemplateData{
        // s.fileOptions,
        template.HTML(string(asJson)),
    })
}

func (s *MashAppServer) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
    asPath := fmt.Sprintf("%s/%s", s.rootPath, templateName)
    t, err := template.ParseFiles(asPath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func serveStaticFiles(fromDirectory string, toHttpPrefix string) {
    asPath := fmt.Sprintf("/%s/", toHttpPrefix)
    fs := http.FileServer(http.Dir(fromDirectory))
    http.Handle(asPath, disableCache(http.StripPrefix(asPath, fs)))
}

func disableCache(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // PICK: Remove ever?
        w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
        w.Header().Set("Pragma", "no-cache") // HTTP 1.0.
        w.Header().Set("Expires", "0") // Proxies.
        h.ServeHTTP(w, r)
    })
}

func listMusicFiles(fromDirectory string) []string {
    infos, err := ioutil.ReadDir(fromDirectory)
    if err != nil {
        panic("Oops, can't read directory")
    }

    result := make([]string, 0)
    for _, info := range infos {
        if !info.IsDir() && isMusicFile(info.Name()) {
            result = append(result, info.Name())
        }
    }
    return result
}

func isMusicFile(name string) bool {
    return strings.HasSuffix(name, ".wav") || strings.HasSuffix(name, ".flac")
}

