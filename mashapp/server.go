package mashapp

import (
    "fmt"
    "html/template"
    "net/http"    
)

type MashAppServer struct {
    port int
    rootPath string
    state *ServerState
}

func NewServer(port int, rootPath string) *MashAppServer {
    return &MashAppServer{
        port, 
        rootPath,
        NewServerState(),
    }
}

func (s *MashAppServer) Serve() {
    addr := fmt.Sprintf(":%d", s.port)
    fmt.Printf("Serving http://localhost%s/\n", addr)

    serveStaticFiles(fmt.Sprintf("%s/static", s.rootPath), "static")

    s.serveRPCs()

    http.HandleFunc("/", s.appHandler)
    
    http.ListenAndServe(addr, nil)
}

func (s *MashAppServer) appHandler(w http.ResponseWriter, r *http.Request) {
    s.renderTemplate(w, "app.html", nil)
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
    http.Handle(asPath, http.StripPrefix(asPath, fs))
}
