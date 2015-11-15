package mashapp

import (
    "encoding/json"
    "fmt"
    "net/http" 
)

func (s *MashAppServer) serveRPCs() {
    s.serveRPC("input/load", s.wrapLoad)
    s.serveRPC("input/edit", s.wrapEdit)
    s.serveRPC("block/new", s.wrapCreateBlock)
}

func (s *MashAppServer) serveRPC(path string, handleFunc func(rw http.ResponseWriter, req *http.Request)) {
    rpcPath := fmt.Sprintf("/_/%s", path)
    fmt.Printf("Adding RPC handler for %s\n", rpcPath)
    http.HandleFunc(rpcPath, handleFunc)
}

// Load input RPC
type LoadRequest struct {
    Path string `json:"path"`
}
type LoadResponse struct {
    Input InputMeta `json:"meta"`
    Samples JsonSamples `json:"samples"`
}
func (s *MashAppServer) wrapLoad(w http.ResponseWriter, r *http.Request) {
    var in LoadRequest
    err := json.NewDecoder(r.Body).Decode(&in)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    out := s.performLoad(in)
    js, err := json.Marshal(out)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}
func (s *MashAppServer) performLoad(req LoadRequest) LoadResponse {
    // TODO: error handling
    id, sound := s.state.loadSound(fmt.Sprintf("%s/%s", s.filePath, req.Path))

    meta := InputMeta{
        id, req.Path, false /* Muted */,
        len(sound.samples), len(sound.samples), /* Duration */
        0, 0, /* Pitch */
    }
    return LoadResponse{meta, floatsToBase64(sound.samples)}
}

// Edit Input RPC
type EditRequest struct {
    Meta InputMeta `json:"meta"`
}
type EditResponse struct {
    Input InputMeta `json:"meta"`
    Samples JsonSamples `json:"samples"`
}
func (s *MashAppServer) wrapEdit(w http.ResponseWriter, r *http.Request) {
    var in EditRequest
    err := json.NewDecoder(r.Body).Decode(&in)
    if err != nil {
        fmt.Printf("Decode error :(\n")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    out := s.performEdit(in)
    js, err := json.Marshal(out)
    if err != nil {
        fmt.Printf("Marshal error :(\n")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}
func (s *MashAppServer) performEdit(req EditRequest) EditResponse {
    // TODO: error handling
    newSamples := s.state.shiftInput(req.Meta).samples
    return EditResponse{req.Meta, floatsToBase64(newSamples)}
}

// Create Block RPC
type CreateBlockRequest struct {
    Block Block `json:"block"`
}
type CreateBlockResponse struct {
    Block Block `json:"block"`
}
func (s *MashAppServer) wrapCreateBlock(w http.ResponseWriter, r *http.Request) {
    var in CreateBlockRequest
    err := json.NewDecoder(r.Body).Decode(&in)
    if err != nil {
        fmt.Printf("Decode error :(\n")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    out := s.performCreateBlock(in)
    js, err := json.Marshal(out)
    if err != nil {
        fmt.Printf("Marshal error :(\n")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
}
func (s *MashAppServer) performCreateBlock(req CreateBlockRequest) CreateBlockResponse {
    block := s.state.createBlock(req.Block)
    return CreateBlockResponse{block}
}