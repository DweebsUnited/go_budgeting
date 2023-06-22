package shiftpath

import (
	"net/http"
	"path"
	"strings"
)

// Taken from here: https://benhoyt.com/writings/go-routing/#shiftpath
//   which references a now dead blog article

// ShiftPath splits the given path into the first segment (head) and
// the rest (tail). For example, "/foo/bar/baz" gives "foo", "/bar/baz".
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

// EnsureMethod is a helper that reports whether the request's method is
// the given method, writing an Allow header and a 405 Method Not Allowed
// if not. The caller should return from the handler if this returns false.
func EnsureMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if method != r.Method {
		w.Header().Set("Allow", method)
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
