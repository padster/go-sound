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

    // TODO - pitch and duration handling
    meta := InputMeta{
        id, req.Path, false /* Muted */,
        len(sound.samples), len(sound.samples), /* Duration */
        0, 0, /* Pitch */
    }

    return LoadResponse{meta, floatsToBase64(sound.samples)}
}

func floatsToBase64(values GoSamples) JsonSamples {
    asFloats := ([]float64)(values)
    return JsonSamples(bytesToBase64(floatsToBytes(asFloats)))
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
