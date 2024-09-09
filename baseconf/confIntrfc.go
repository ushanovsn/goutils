package baseconf

// Interface for use the configuration pack. GetConfigUniversalPtr() - returning pointer (changed type to interface{}) to configuration structure.
//
// The configuration structure should have tags "cfg" (with name of parameter in conf file w/o spaces) and "descr" (with descriptions\comments for parameter)
type UseConf interface {
	// Getting the config file path (or just name)
	GetConfFileName() string
	// Getting the pointer to configurations structure
	GetConfigUniversalPtr() interface{}
	// Getting the description (common comments) for header of config file
	GetConfigDescr() string
}

// Methods for a compatible logger
type ConfLogging interface {
	// Write log w logging level = "Debug"
	Debug(string)
	// Write log w logging level = "Info"
	Info(string)
	// Write log w logging level = "Warning"
	Warn(string)
	// Write log w logging level = "Error"
	Error(string)
}
