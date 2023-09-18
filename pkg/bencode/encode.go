package bencode

// For the encoding operations, the ready-made library will be used
import (
	"bytes"

	"github.com/jackpal/bencode-go"
)

func EncodeBencodeToString(data interface{}) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	err := bencode.Marshal(buf, data)
	return buf.String(), err
}
