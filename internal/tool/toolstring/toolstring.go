// Package toolstring string tool
package toolstring

import (
	"net/url"
	"regexp"
	"strings"
)

// IsTrue check if string is true
func IsTrue(value string) bool {

	return value == "1" || strings.ToLower(value) == "true"
}

// Precompiled regular expression for reuse
var regexVarName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// IsValidID checks if the given id is a valid identifier
func IsValidID(id string) bool {
	// Ensure id is not empty and matches the regex pattern
	return id != "" && regexVarName.MatchString(id)
}

// LocalURL join like path?args[0]=args[1]&args[2]=args[3]#args[4]
// ignore query key-value pair or fragment if is empty
func LocalURL(path string, args ...string) string {

	count := len(args)
	pairs := count / 2

	if pairs > 0 {

		u := url.URL{
			Path: path,
		}
		query := u.Query()
		for i := 0; i < pairs; i++ {
			k := args[i*2]
			v := args[i*2+1]
			if k != "" && v != "" {
				query.Add(k, v) // this not keep order, internal sort by key on encode
			}

		}

		u.RawQuery = query.Encode()

		if count%2 == 1 {
			v := args[count-1]
			if v != "" {
				u.Fragment = args[count-1]
			}
		}

		return u.String()

	}

	return path
}
