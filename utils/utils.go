package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/olebedev/config"
)

const (
	SimpleDateFormat    = "Jan 2"
	SimpleTimeFormat    = "15:04 MST"
	MinimumTimeFormat12 = "3:04 PM"
	MinimumTimeFormat24 = "15:04"

	FullDateFormat         = "Monday, Jan 2"
	FriendlyDateFormat     = "Mon, Jan 2"
	FriendlyDateTimeFormat = "Mon, Jan 2, 15:04"

	TimestampFormat = "2006-01-02T15:04:05-0700"
)

// DoesNotInclude takes a slice of strings and a target string and returns
// TRUE if the slice does not include the target, FALSE if it does
//
// Example:
//
//    x := DoesNotInclude([]string{"cat", "dog", "rat"}, "dog")
//    > false
//
//    x := DoesNotInclude([]string{"cat", "dog", "rat"}, "pig")
//    > true
//
func DoesNotInclude(strs []string, val string) bool {
	return !Includes(strs, val)
}

// ExecuteCommand executes an external command on the local machine as the current user
func ExecuteCommand(cmd *exec.Cmd) string {
	if cmd == nil {
		return ""
	}

	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return err.Error()
	}

	return buf.String()
}

// FindMatch takes a regex pattern and a string of data and returns back all the matches
// in that string
func FindMatch(pattern string, data string) [][]string {
	r := regexp.MustCompile(pattern)
	return r.FindAllStringSubmatch(data, -1)
}

// Includes takes a slice of strings and a target string and returns
// TRUE if the slice includes the target, FALSE if it does not
//
// Example:
//
//    x := Includes([]string{"cat", "dog", "rat"}, "dog")
//    > true
//
//    x := Includes([]string{"cat", "dog", "rat"}, "pig")
//    > false
//
func Includes(strs []string, val string) bool {
	for _, str := range strs {
		if val == str {
			return true
		}
	}
	return false
}

// OpenFile opens the file defined in `path` via the operating system
func OpenFile(path string) {
	if (strings.HasPrefix(path, "http://")) || (strings.HasPrefix(path, "https://")) {
		if len(OpenUrlUtil) > 0 {
			commands := append(OpenUrlUtil, path)
			args := commands[1:len(commands)]
			exec.Command(commands[0], args...).Start()
			return
		}
		switch runtime.GOOS {
		case "linux":
			exec.Command("xdg-open", path).Start()
		case "windows":
			exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
		case "darwin":
			exec.Command("open", path).Start()
		default:
			// for the BSDs
			exec.Command("xdg-open", path).Start()
		}
	} else {
		filePath, _ := ExpandHomeDir(path)
		cmd := exec.Command(OpenFileUtil, filePath)
		ExecuteCommand(cmd)
	}
}

// ReadFileBytes reads the contents of a file and returns those contents as a slice of bytes
func ReadFileBytes(filePath string) ([]byte, error) {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []byte{}, err
	}

	return fileData, nil
}

// ParseJSON is a standard JSON reader from text
func ParseJSON(obj interface{}, text io.Reader) error {
	d := json.NewDecoder(text)
	return d.Decode(obj)
}

// CalculateDimensions reads the module dimensions from the module and global config. The border is already substracted.
func CalculateDimensions(moduleConfig, globalConfig *config.Config) (int, int) {
	// Read the source data from the config
	left := moduleConfig.UInt("position.left", 0)
	top := moduleConfig.UInt("position.top", 0)
	width := moduleConfig.UInt("position.width", 0)
	height := moduleConfig.UInt("position.height", 0)

	cols := ToInts(globalConfig.UList("wtf.grid.columns"))
	rows := ToInts(globalConfig.UList("wtf.grid.rows"))

	// Make sure the values are in bounds
	left = Clamp(left, 0, len(cols)-1)
	top = Clamp(top, 0, len(rows)-1)
	width = Clamp(width, 0, len(cols)-left)
	height = Clamp(height, 0, len(rows)-top)

	// Start with the border subtracted and add all the spanned rows and cols
	w, h := -2, -2
	for _, x := range cols[left : left+width] {
		w += x
	}
	for _, y := range rows[top : top+height] {
		h += y
	}

	// The usable space may be empty
	w = MaxInt(w, 0)
	h = MaxInt(h, 0)

	return w, h
}

// MaxInt returns the larger of x or y
//
// Examples:
//
//   MaxInt(3, 2) => 3
//   MaxInt(2, 3) => 3
//
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Clamp restricts values to a minimum and maximum value
//
// Examples:
//
//   clamp(6, 3, 8) => 4
//   clamp(1, 3, 8) => 3
//   clamp(9, 3, 8) => 8
//
func Clamp(x, a, b int) int {
	if a > x {
		return a
	}
	if b < x {
		return b
	}
	return x
}
