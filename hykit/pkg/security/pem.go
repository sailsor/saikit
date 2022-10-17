package security

import (
	"bytes"
	"encoding/pem"

	"github.com/pkg/errors"
)

var pemStart = []byte("\n-----BEGIN ")
var pemEnd = []byte("\n-----END ")
var pemEndOfLine = []byte("-----")

func preloadPEM(data []byte) (rest []byte, err error) {
	// pemStart begins with a newline. However, at the very beginning of
	// the byte array, we'll accept the start string without it.
	rest = data
	if bytes.HasPrefix(data, pemStart[1:]) {
		rest = rest[len(pemStart)-1 : len(data)]
	} else {
		return nil, errors.New("非法PEM格式")
	}

	if bytes.Contains(rest, pemEndOfLine) {
		rest = rest[bytes.Index(rest, pemEndOfLine)+len(pemEndOfLine):]
	} else {
		return nil, errors.New("非法PEM格式")
	}

	for i := 0; i < len(rest); i++ {
		p := rest[i:len(rest)]
		if bytes.HasPrefix(p, pemEnd[1:]) {
			break
		}
		if rest[i] == 0x20 {
			rest[i] = '\n'
		}
	}
	return data, nil
}

func DecodePEM(data []byte) (*pem.Block, error) {
	pemBuf, err := preloadPEM(data)
	if err != nil {
		return nil, err
	}

	p, rest := pem.Decode(pemBuf)
	if len(p.Bytes) == 0 {
		return nil, errors.New("PEM Decode Error")
	}
	if len(rest) != 0 {
		return nil, errors.Errorf("PEM Decode Left Error[%v]", rest)
	}

	return p, nil
}
