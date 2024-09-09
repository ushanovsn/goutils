package baseconf

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Processing and initializing config file. The logger should implement the interface or be nil.
//
// The configuration structure should have public/exported parameters and tags: "cfg" (with name of parameter in conf file w/o spaces) and "descr" (with any descriptions\comments for parameter).
// If there are no tags - parameter will be ignoring and not set into the file.
// File with configuration can be exists or it will be created using tags and values, the structure have.
func ProcConfig(cfg UseConf, log ConfLogging) bool {
	if log == nil {
		log = &elog{}
	}
	fPath := cfg.GetConfFileName()

	if wd, err := os.Getwd(); err == nil {
		log.Info(fmt.Sprintf("Current directory: \"%s\"", wd))
	} else {
		log.Error(fmt.Sprintf("Error in computing the current directory: \"%s\"", err.Error()))
	}
	log.Info(fmt.Sprintf("Start to load configuration from \"%s\" file", fPath))

	// check the file
	if _, err := os.Stat(fPath); err == nil {
		// file exist, now get config data
		err := readConfig(fPath, cfg.GetConfigUniversalPtr(), log)
		if err != nil {
			log.Error("Error while read config file")
			return false
		}

	} else if errors.Is(err, os.ErrNotExist) {
		// file is not exist, need create it
		log.Warn("Config file is not exist")
		if !createConfigFile(fPath, cfg.GetConfigUniversalPtr(), cfg.GetConfigDescr(), log) {
			log.Error("Error occured when Config file creating")
			return false
		}
	} else {
		log.Error("Error while checking config file: " + err.Error())
		return false
	}

	log.Info("Config file was processed")
	return true
}

// Read and apply config from file
func readConfig(filePath string, config interface{}, log ConfLogging) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Info("Reading config file started")

	// readed error lines count
	var errCnt int
	// reflect value of input config structure
	var confRef = reflect.ValueOf(config).Elem()

	// scan file by lines
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// skip empty lines and comments
		if line == "" || line[0] == '#' {
			continue
		}

		// split parameter string
		words := strings.Split(line, "=")

		// correct string contains one symbol "=", Name of parameter at left part and Value at right part
		if len(words) != 2 {
			errCnt++
			log.Debug(fmt.Sprintf("Incorrect string was readed: \"%s\"", line))
			continue
		}

		readName := strings.TrimSpace(words[0])
		readVal := strings.TrimSpace(words[1])

		// trying to find readed parameter in config structure by tags
		for i := 0; i < confRef.NumField(); i++ {
			// name of parameter in tag "cfg"
			if readName == confRef.Type().Field(i).Tag.Get("cfg") {
				log.Debug(fmt.Sprintf("Parameter \"%s\" type (%v) match, now assign value: %s", readName, (confRef.Field(i).Type()), readVal))
				// check access to parameter in structure
				if !confRef.Field(i).CanSet() {
					log.Error(fmt.Sprintf("Missing acess to parameter \"%s\"", readName))
					break
				}

				// parsing error
				var convErr error
				// bad access for use type flag
				var cantUse bool
				// switch needed type of value
				switch confRef.Field(i).Interface().(type) {
				case int:
					if confRef.Field(i).CanInt() {
						v, err := strconv.Atoi(readVal)
						if err == nil {
							confRef.Field(i).SetInt(int64(v))
						} else {
							convErr = err
						}
					} else {
						cantUse = true
					}
				case uint:
					if confRef.Field(i).CanUint() {
						v, err := strconv.ParseUint(readVal, 10, 64)
						if err == nil {
							confRef.Field(i).SetUint(v)
						} else {
							convErr = err
						}
					} else {
						cantUse = true
					}
				case float64:
					if confRef.Field(i).CanFloat() {
						v, err := strconv.ParseFloat(readVal, 64)
						if err == nil {
							confRef.Field(i).SetFloat(v)
						} else {
							convErr = err
						}
					} else {
						cantUse = true
					}
				case bool:
					// for bool enough check for CanSet()
					v, err := strconv.ParseBool(readVal)
					if err == nil {
						confRef.Field(i).SetBool(v)
					} else {
						convErr = err
					}
				case string:
					// string don't need conversion and it enough check for CanSet()
					confRef.Field(i).SetString(readVal)
				default:
					log.Error(fmt.Sprintf("Type \"%s\" of parameter \"%s\" was skipped while processing config file", (confRef.Field(i).Type()), readName))
				}

				// logging errors
				if cantUse {
					log.Error(fmt.Sprintf("Parameter \"%s\". Can't use type: %v", readName, (confRef.Field(i).Type())))
				}
				if convErr != nil {
					log.Error(fmt.Sprintf("Parameter \"%s\", error when converse string \"%s\" to %s. Error: %s", readName, readVal, (confRef.Field(i).Type()), convErr.Error()))
				}

				// now parameter is found and processed
				break
			}
		}
	}

	// error when scan the file
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// creating config file w header text
func createConfigFile(filePath string, config interface{}, descr string, log ConfLogging) bool {
	fRes := true
	headerConf := `###   ` + descr + `   ###` + "\n" +
		"#\n" +
		`# Config file contains "Name" and "Value" of parameters separated by a symbol "="` + "\n" +
		`# Symbol "=" is allowed to use no more than 1 piece per line` + "\n" +
		`# If this file is deleted, the service will automatically create a new file at startup` + "\n" +
		`# File will be filled with all valid parameters with default values` + "\n" +
		`#` + "\n" +
		`# Comments should start with the "#" character from the beginning of the line` + "\n" +
		"\n\n"

	log.Info("Creating config file...")

	// create the file
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Error("Error while creating config file. Err: " + err.Error())
		fRes = false
	}
	defer f.Close()

	// write header into the file
	if _, err = f.WriteString(headerConf); err != nil {
		log.Error("Error while write data in config file. Err: " + err.Error())
		fRes = false
	}

	// reflect value of input config structure
	var confRef = reflect.ValueOf(config).Elem()

	// filling the file by parameters of configuration
	for i := 0; i < confRef.NumField(); i++ {
		// name of parameter in tag "cfg"
		pName := confRef.Type().Field(i).Tag.Get("cfg")
		// description of parameter in tag "descr"
		pDescr := confRef.Type().Field(i).Tag.Get("descr")
		// value of parameter
		pValue := confRef.Field(i)

		if pName != "" {
			lines := "\n" + "# " + pDescr + "\n" + fmt.Sprintf("%s = %v", pName, pValue) + "\n"

			if _, err = f.WriteString(lines); err != nil {
				log.Debug("Error while write lines: \n" + lines)
				log.Error("Error while write line in config file. Err: " + err.Error())
				fRes = false
			}
		}
	}

	log.Info("Config file was created and filled defaults values")

	return fRes
}
