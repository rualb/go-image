// Package toolconfig ...
package toolconfig

import (
	"encoding/json"
	"fmt"
	"go-image/internal/tool/toolhttp"
	xlog "go-image/internal/tool/toollog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func LoadConfig(cnfPtr any, dir string, fileName string) error {

	xlog.Info("Loading config from: %v", dir)

	isHTTP := strings.HasPrefix(dir, "http")

	// TODO from s3://

	if isHTTP {

		err := fromURL(cnfPtr, dir, fileName)
		if err != nil {
			return err
		}

	} else {
		err := fromFile(cnfPtr, dir, fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// FromFile errIfNotExists argument soft binding, no error if file not exists
func fromFile(cnfPtr any, dir string, file string) error {

	if file == "" {
		return nil
	}

	if !strings.HasSuffix(file, ".json") {
		return fmt.Errorf("error file not match  *.json: %v", file)
	}

	fullPath, err := filepath.Abs(filepath.Join(dir, file))

	if err != nil {
		return err
	}

	fullPath = filepath.Clean(fullPath)

	data, err := os.ReadFile(fullPath)

	if err != nil {
		return fmt.Errorf("error with file %v: %v", fullPath, err)
	}

	xlog.Info("Loading config from file: %v", fullPath)

	err = fromJSON(cnfPtr, string(data))

	if err != nil {
		return err
	}

	return nil
}

// FromURL errIfNotExists argument soft binding, no error if file not exists
func fromURL(cnfPtr any, dir string, file string) error {

	if file == "" {
		return nil
	}

	if !strings.HasSuffix(file, ".json") {
		return fmt.Errorf("error file not match  *.json: %v", file)
	}

	fullPath := dir + "/" + file

	_, err := url.Parse(fullPath)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// fmt.Println("Reading config from file: ", file)

	data, err := toolhttp.GetBytes(fullPath, nil, nil)

	if err != nil {
		return fmt.Errorf("error with file %v: %v", fullPath, err)
	}

	xlog.Info("Loading config from file: %v", fullPath)

	err = fromJSON(cnfPtr, string(data))
	if err != nil {
		return err
	}

	return nil
}

func expandEnv(data string) string {

	re := regexp.MustCompile(`\$\{([A-Z_][0-9_A-Z]*)\}`)
	data = re.ReplaceAllStringFunc(data, func(match string) string {
		name := match[2 : len(match)-1] // Remove '${' and '}'
		val := os.Getenv(name)
		if val == "" {
			xlog.Warn("Missing env value for: %v", match)
		}
		return val // Return the original match if not found
	})

	return data

	// data = os.Expand(data, func(s string) string {

	// 	// TODO chek if var name [A-Z_0-9]+

	// 	parts := strings.SplitN(s, ":", 2)
	// 	name := parts[0]
	// 	// tail:=parts[1]
	// 	val := os.Getenv(name)

	// 	if val == "" {
	// 		//
	// 		xlog.Warn("Missing env value for: %v", s)
	// 	}

	// 	return val

	// })

}

func fromJSON(cnfPtr any, data string) error {

	if data == "" {
		return nil
	}

	data = expandEnv(data)

	err := json.Unmarshal([]byte(data), cnfPtr)

	if err != nil {
		return err
	}

	return nil
}
