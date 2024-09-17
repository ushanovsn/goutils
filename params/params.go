package params

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Enum types of holding parameters
type DataType int

const (
	DataBase64 DataType = iota
	DataZip
	DataEncrypt
)

// Name of parameter that holding types of holding
const defTypeParamName = "CurrentDataTypeForParameters"

// Parameters Object for interaction
//
// Params read from file when Get() and write into file when Set(), i.e. file not blocked all time.
// This pack for slow access and careful storage. Encryption option is available.
// First line in the file is DataType parameter.
type ParamsObj struct {
	fileName   string
	wDir       string
	encryptKey string
	typeData   DataType
	mtx        sync.RWMutex
}

// Init new ParamObject
//
// fName - file name or path; t - type storages; pass - password when data is stored encrypted.
// Uses existing param file or automatically creates new.
func New(fName string, t DataType, pass string) (*ParamsObj, error) {
	// create base object
	p := ParamsObj{
		fileName:   fName,
		typeData:   t,
		encryptKey: pass,
	}

	// check file name
	if fName == "" {
		return nil, fmt.Errorf("Empty filename receiving")
	}

	// check current work directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error in computing the current directory: \"%s\"", err.Error())
	}
	p.wDir = wd

	// check file access
	if _, err := os.Stat(fName); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file is not exist, need create it
			f, err := os.Create(fName)
			if err != nil {
				return nil, fmt.Errorf("Error when creating param file: \"%s\"", err.Error())
			}
			f.Close()
			// write first line with type
			err = p.SetValue(defTypeParamName, fmt.Sprintf("%v", t))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Access file error: \"%s\"", err.Error())
		}
	}

	// Read first line with DataType and check current DataType parameter
	if val, err := p.GetValue(defTypeParamName); err != nil || !checkDataType(val, p.typeData) {
		return nil, fmt.Errorf("Wrong DataType for existing file. Error: \"%s\"", err.Error())
	}

	return &p, nil
}

// Writing a parameter to a file. The name starting with letter, can consist of letters, numbers, and symbols "-", "_".
func (obj *ParamsObj) SetValue(n string, val string) error {
	// check incoming value
	if val == "" {
		return fmt.Errorf("Empty value for writing parameter")
	}

	// check correct name
	if err := checkName(n); !err {
		return fmt.Errorf("Wrong Name parameter: \"%s\"", n)
	}

	// now write parameter into file
	err := obj.writeValue(n, val)
	if err != nil {
		return fmt.Errorf("Error writing parameter into file:%s ", err.Error())
	}

	return nil
}

// Reading a parameter from a file. The name starting with letter, can consist of letters, numbers, and symbols "-", "_".
func (obj *ParamsObj) GetValue(n string) (val string, err error) {
	// check correct name
	if err := checkName(n); !err {
		return val, fmt.Errorf("Wrong Name parameter: \"%s\"", n)
	}

	// now read parameter
	return obj.readValue(n)
}

// Deleting parameter from file
func (obj *ParamsObj) DeleteValue(n string) error {
	return obj.deleteValue(n)
}

// Check value in string vs DataType
func checkDataType(v string, t DataType) bool {
	// convert and check errors
	if conv, err := strconv.Atoi(v); err == nil && conv == int(t) {
		return true
	}

	return false
}

// Name not empty starting with letter and consists of letters, numbers, and symbols "-", "_"
func checkName(n string) bool {
	ok, err := regexp.Match(`^[\w][\w_-]*\z`, []byte(n))
	return ok && err == nil
}

// Write value into file, rewrite if exist
func (obj *ParamsObj) writeValue(n string, val string) error {
	// block mutex for all period of writing file
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	// open file with params
	f, err := os.OpenFile(obj.fileName, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	var foundFlag bool = false

	enc, err := obj.packValue(val)
	if err != nil {
		return err
	}

	// scan file by lines
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// check parameter name
		if ok, err := regexp.Match(`^`+n+`\$.*`, []byte(line)); !(ok && err == nil) {
			lines = append(lines, line)
		} else { // match name!!! update it
			lines = append(lines, n+"$"+enc)
			foundFlag = true
		}
	}

	// error when scan the file
	if err := scanner.Err(); err != nil {
		return err
	}

	if !foundFlag {
		lines = append(lines, n+"$"+enc)
	}

	output := strings.Join(lines, "\n")

	// clear old data
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("Error when truncate params file. Err: %s", err.Error())
	}

	// set cursor at start positiom
	_, err = f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("Error when move cursor for params file. Err: %s", err.Error())
	}

	// write new file
	_, err = f.Write([]byte(output))

	if err != nil {
		return err
	}

	return nil
}

// Reading file, searching by name and return encoding value. Error when not found.
func (obj *ParamsObj) readValue(n string) (val string, err error) {
	// block mutex for all period of reading file
	obj.mtx.RLock()
	defer obj.mtx.RUnlock()

	// open file with params
	f, err := os.Open(obj.fileName)
	if err != nil {
		return val, err
	}
	defer f.Close()

	// scan file by lines
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// skip empty lines
		if line == "" {
			continue
		}

		// find parameter name
		if ok, err := regexp.Match(`^`+n+`\$.*`, []byte(line)); !(ok && err == nil) {
			// skip not matched line
			continue
		} else { // match name!!!
			// extract value
			re, _ := regexp.Compile(`^` + n + `\$(.*)`)
			mVal := re.FindStringSubmatch(line)
			if mVal == nil || len(mVal) < 2 || (len(mVal) > 1 && mVal[1] == "") {
				return val, fmt.Errorf("Parameter \"%s\" was found, but value is empty", n)
			}

			// unpack or decoding
			return obj.unpackValue(mVal[1])
		}
	}

	// error when scan the file
	if err := scanner.Err(); err != nil {
		return val, err
	}

	return val, fmt.Errorf("Parameter \"%s\" not found", n)
}

// Check current DataType and perfom coversion data
func (obj *ParamsObj) unpackValue(rawVal string) (val string, err error) {

	switch obj.typeData {
	case DataZip:
		b, err := base64.StdEncoding.DecodeString(rawVal)
		if err != nil {
			return "", err
		}
		data, err := decompress(b)
		return string(data), err
	case DataBase64:
		b, err := base64.StdEncoding.DecodeString(rawVal)
		return string(b), err
	case DataEncrypt:
		b, err := base64.StdEncoding.DecodeString(rawVal)
		if err != nil {
			return "", err
		}
		data, err := decrypt(string(b), obj.encryptKey)
		return string(data), err
	default:
		return rawVal, err
	}
}

// Check current DataType and perfom decoversion data
func (obj *ParamsObj) packValue(rawVal string) (val string, err error) {
	var outB64 string

	switch obj.typeData {
	case DataZip:
		data, err := compress([]byte(rawVal))
		if err != nil {
			return "", err
		}
		outB64 = base64.StdEncoding.EncodeToString(data)
		return outB64, nil
	case DataBase64:
		return base64.StdEncoding.EncodeToString([]byte(rawVal)), err
	case DataEncrypt:
		data, err := encrypt(rawVal, obj.encryptKey)
		if err != nil {
			return "", err
		}
		outB64 = base64.StdEncoding.EncodeToString(data)
		return outB64, nil
	default:
		return rawVal, err
	}
}

// Delete value in file
func (obj *ParamsObj) deleteValue(n string) error {
	// block mutex for all period of using file
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	// open file with params
	f, err := os.OpenFile(obj.fileName, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string

	// scan file by lines
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// check parameter name and memorize only not matching
		if ok, err := regexp.Match(`^`+n+`\$.*`, []byte(line)); !(ok && err == nil) {
			lines = append(lines, line)
		}
	}

	// error when scan the file
	if err := scanner.Err(); err != nil {
		return err
	}

	output := strings.Join(lines, "\n")

	// clear old data
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("Error when truncate params file. Err: %s", err.Error())
	}

	// set cursor at start positiom
	_, err = f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("Error when move cursor for params file. Err: %s", err.Error())
	}

	// write new file using corrected data
	_, err = f.Write([]byte(output))

	if err != nil {
		return err
	}

	return nil
}
