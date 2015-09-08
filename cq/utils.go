package cq

import (
  "encoding/binary"
  "io"
)

func WriteComplexBlock(w io.Writer, block [][]complex128) {
  WriteInt32(w, int32(len(block)))
  for _, b := range block {
    WriteComplexArray(w, b)
  }
}

func WriteComplexArray(w io.Writer, array []complex128) {
  WriteInt32(w, int32(len(array)))
  for _, c := range array {
    WriteComplex(w, c)
  }
}

func WriteComplex(w io.Writer, c complex128) {
  WriteFloat32(w, float32(real(c)))
  WriteFloat32(w, float32(imag(c)))
}

func WriteInt32(w io.Writer, i int32) {
  binary.Write(w, binary.LittleEndian, i)
}

func WriteFloat32(w io.Writer, f float32) {
  binary.Write(w, binary.LittleEndian, f)
}
