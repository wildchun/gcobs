package gcobs

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	encodeData := make([][]byte, 0)
	for i := 1; i < 10240; i += 100 {
		tmp := make([]byte, 0, i)
		for j := 0; j < i; j++ {
			tmp = append(tmp, byte(j))
		}
		encodeData = append(encodeData, tmp)
	}
	for i := 0; i < len(encodeData); i++ {
		encoded, err := Encode(encodeData[i])
		if err != nil {
			t.Error("Error encoding data", err)
		}
		decoded, err := Decode(encoded)
		if err != nil {
			t.Error("Error decoding data: ", err)
		}
		if !bytes.Equal(encodeData[i], decoded) {
			t.Error("Encoded and decoded data do not match")
		}
	}
}
