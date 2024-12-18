package coswss

import (
	"github.com/hwcer/cosnet"
	"net/http"
)

var Options = struct {
	Accept func(s *cosnet.Socket, meta map[string]string)
	Verify func(w http.ResponseWriter, r *http.Request) (meta map[string]string, err error)
	Origin []string
}{}
