package mashapp

import (
    b64 "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "math"
    "net/http" 
)

func (s *MashAppServer) serveRPCs() {
    s.serveRPC("load", s.wrapLoad)
}
func (s *MashAppServer) serveRPC(path string, handleFunc func(rw http.ResponseWriter, req *http.Request)) {
    rpcPath := fmt.Sprintf("/_/%s", path)
    http.HandleFunc(rpcPath, handleFunc)
}

type LoadInput struct {
    Path string `json:"path"`
}
type LoadOutput struct {
    ID int `json:"path"`
    Samples string `json:"samples"`
}
func (s *MashAppServer) wrapLoad(w http.ResponseWriter, r *http.Request) {
    var in LoadInput
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
func (s *MashAppServer) performLoad(in LoadInput) LoadOutput {
    // TODO: error handling
    id, sound := s.state.loadSound(in.Path)

    return LoadOutput{
        id,
        floatsToBase64(sound.samples),
    }
}

func floatsToBase64(values []float64) string {
    return bytesToBase64(floatsToBytes(values))
}

func floatsToBytes(values []float64) []byte {
    bytes := make([]byte, 0, 4 * len(values))
    for _, v := range values {
        bytes = append(bytes, float32ToBytes(float32(v))...)
    }
    return bytes
}

func float32ToBytes(value float32) []byte {
    bits := math.Float32bits(value)
    bytes := make([]byte, 4)
    binary.LittleEndian.PutUint32(bytes, bits)
    return bytes
}

func bytesToBase64(values []byte) string {
    return b64.StdEncoding.EncodeToString(values)
}