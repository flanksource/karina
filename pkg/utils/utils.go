package utils

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	reset        = "\x1b[0m"
	red          = "\x1b[31m"
	lightRed     = "\x1b[31;1m"
	green        = "\x1b[32m"
	lightGreen   = "\x1b[32;1m"
	lightBlue    = "\x1b[34;1m"
	magenta      = "\x1b[35m"
	lightMagenta = "\x1b[35;1m"
	cyan         = "\x1b[36m"
	lightCyan    = "\x1b[36;1m"
	white        = "\x1b[37;1m"
	bold         = "\x1b[1m"
	boldOff      = "\x1b[22m"
)

//SafeExec executes the sh script and returns the stdout and stderr, errors will result in a nil return only.
func SafeExec(sh string, args ...interface{}) (string, bool) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(sh, args...))
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Debugf("Failed to exec %s, %s %s\n", sh, data, err)
		return "", false
	}

	if !cmd.ProcessState.Success() {
		log.Debugf("Command did not succeed %s\n", sh)
		return "", false
	}
	return string(data), true

}

//Exec runs the sh script and forwards stderr/stdout to the console
func Exec(sh string, args ...interface{}) error {
	log.Debugf("exec: "+sh, args...)
	cmd := exec.Command("bash", "-c", fmt.Sprintf(sh, args...))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s failed with %s", sh, err)
	}

	if !cmd.ProcessState.Success() {
		return fmt.Errorf("%s failed to run", sh)
	}
	return nil
}

func takeSliceArg(arg interface{}) (out []interface{}, ok bool) {
	val := reflect.ValueOf(arg)
	if val.Kind() != reflect.Slice {
		return nil, false
	}

	c := val.Len()
	out = make([]interface{}, c)
	for i := 0; i < val.Len(); i++ {
		out[i] = val.Index(i).Interface()
	}
	return out, true
}

//IsSlice returns true if the argument is a slice
func IsSlice(arg interface{}) bool {
	return reflect.ValueOf(arg).Kind() == reflect.Slice
}

//ToString takes an object and tries to convert it to a string
func ToString(i interface{}) string {
	if slice, ok := takeSliceArg(i); ok {
		s := ""
		for _, v := range slice {
			if s != "" {
				s += ", "
			}
			s += ToString(v)
		}
		return s

	}
	switch v := i.(type) {
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	case interface{}:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		// panic(fmt.Sprintf("I don't know about type %T!\n", v))
	}
	return ""
}

//StructToMap takes an object and returns all it's field in a map
func StructToMap(s interface{}) map[string]interface{} {
	values := make(map[string]interface{})
	value := reflect.ValueOf(s)

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if field.CanInterface() {
			v := field.Interface()
			if v != nil && v != "" {
				values[value.Type().Field(i).Name] = v
			}
		}
	}
	return values
}

//StructToIni takes an object and serializes it's fields in INI format
func StructToIni(s interface{}) string {
	str := ""
	for k, v := range StructToMap(s) {
		str += k + "=" + ToString(v) + "\n"
	}
	return str
}

//MapToIni takes a map and converts it into an INI formatted string
func MapToIni(Map map[string]string) string {
	str := ""
	for k, v := range Map {
		str += k + "=" + ToString(v) + "\n"
	}
	return str
}

//IniToMap takes the path to an INI formatted file and transforms it into a map
func IniToMap(path string) map[string]string {
	result := make(map[string]string)
	ini := SafeRead(path)
	for _, line := range strings.Split(ini, "\n") {
		values := strings.Split(line, "=")
		if len(values) == 2 {
			result[values[0]] = values[1]
		}
	}
	return result
}

//GzipFile takes the path to a file and returns a Gzip comppressed byte slic
func GzipFile(path string) ([]byte, error) {
	var buf bytes.Buffer

	w := gzip.NewWriter(&buf)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(contents)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	result := buf.Bytes()
	return result, nil
}

//SafeRead reads a path and returns the text contents or nil,
func SafeRead(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

//ReplaceAllInSlice runs strings.Replace on all elements in a slice and returns the result
func ReplaceAllInSlice(a []string, find string, replacement string) (replaced []string) {
	for _, s := range a {
		replaced = append(replaced, strings.Replace(s, find, replacement, -1))
	}
	return
}

//SplitAllInSlice runs strings.Split on all elements in a slice and returns the results at the given index
func SplitAllInSlice(a []string, split string, index int) (replaced []string) {
	for _, s := range a {
		replaced = append(replaced, strings.Split(s, split)[index])
	}
	return
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetBaseName(filename string) string {
	filename = path.Base(filename)
	parts := strings.Split(filename, ".")
	if len(parts) == 1 {
		return filename
	}
	return strings.Join(parts[0:len(parts)-1], ".")
}

func GetEnvOrDefault(names ...string) string {
	for _, name := range names {
		if val := os.Getenv(name); val != "" {
			return val
		}
	}
	return ""
}

func FileCopy(src string, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func IsTTY() bool {
	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	} else {
		return true
	}
}

func Redf(msg string, args ...interface{}) string {
	if IsTTY() {
		return red + fmt.Sprintf(msg, args...) + reset
	}
	return fmt.Sprintf(msg, args...)
}

func Greenf(msg string, args ...interface{}) string {
	if IsTTY() {
		return green + fmt.Sprintf(msg, args...) + reset
	}
	return fmt.Sprintf(msg, args...)
}

func LightGreenf(msg string, args ...interface{}) string {
	if IsTTY() {
		return lightGreen + fmt.Sprintf(msg, args...) + reset
	}
	return fmt.Sprintf(msg, args...)
}

func LightCyanf(msg string, args ...interface{}) string {
	if IsTTY() {
		return lightCyan + fmt.Sprintf(msg, args...) + reset
	}
	return fmt.Sprintf(msg, args...)
}

// ShortTimestamp returns a shortened timestamp using
// week of year + day of week to represent a day of the
// e.g. 1st of Jan on a Tuesday is 13
func ShortTimestamp() string {
	_, week := time.Now().ISOWeek()
	return fmt.Sprintf("%d%d-%s", week, time.Now().Weekday(), time.Now().Format("150405"))
}

func Interpolate(arg string, vars interface{}) string {
	tmpl, err := template.New("test").Parse(arg)
	if err != nil {
		log.Errorf("Failed to parse template %s -> %s\n", arg, err)
		return arg
	}
	buf := bytes.NewBufferString("")

	err = tmpl.Execute(buf, vars)
	if err != nil {
		log.Errorf("Failed to execute template %s -> %s\n", arg, err)
		return arg
	}
	return buf.String()

}
func InterpolateStrings(arg []string, vars interface{}) []string {
	out := make([]string, len(arg))
	for i, e := range arg {
		out[i] = Interpolate(e, vars)
	}
	return out
}

func ToGenericMap(m map[string]string) map[string]interface{} {
	var out = map[string]interface{}{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func ToStringMap(m map[string]interface{}) map[string]string {
	var out = make(map[string]string)
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func GET(url string, args ...interface{}) ([]byte, error) {
	url = fmt.Sprintf(url, args...)

	resp, err := http.Get(url)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, nil
}

func Download(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// randomChars defines the alphanumeric characters that can be part of a random string
const randomChars = "0123456789abcdefghijklmnopqrstuvwxyz"

// RandString returns a random string consisting of the characters in
// randomChars, with the length customized by the parameter
func RandomString(length int) string {
	// len("0123456789abcdefghijklmnopqrstuvwxyz") = 36 which doesn't evenly divide
	// the possible values of a byte: 256 mod 36 = 4. Discard any random bytes we
	// read that are >= 252 so the bytes we evenly divide the character set.
	const maxByteValue = 252

	var (
		b     byte
		err   error
		token = make([]byte, length)
	)

	reader := bufio.NewReaderSize(rand.Reader, length*2)
	for i := range token {
		for {
			if b, err = reader.ReadByte(); err != nil {
				return ""
			}
			if b < maxByteValue {
				break
			}
		}

		token[i] = randomChars[int(b)%len(randomChars)]
	}

	return string(token)
}
