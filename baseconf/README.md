# baseconf

Package for using base simple text file with configuration.
File consists pairs "key = value".

To use pack, is enough to pass the structure (which describes the configuration), structure parameters types and string values must match and be convertible. .
The configuration structure should have public/exported parameters and this tags:
* "cfg" - with name of parameter in conf file (w/o spaces)
* "descr" - with any descriptions/comments for parameter
Script use parameters and will read the config file and associate config struct or create new file using tags.

</br>

Added simple interfaces for incoming objects:

1. UseConf interface

    * Must return file name for configuration (at work dir) or full path:
    ```GO
    GetConfFileName() string
    ```

    * Must return pointer to configuration structure with tags:
    ```GO
    GetConfigUniversalPtr() interface{}
    ```
    like this:
    ```GO
    func (obj *<T1>) GetConfigUniversalPtr() interface{} {
	    return &<T1>.configStructureAnyType
    }
    ```

    * Must return description (any text) to place it as comment into the "header" of file configuration:
    ```GO
    GetConfigDescr() string
    ```

2. ConfLogging interface

    * Most frequently used methods for logging of most popular loggers (logrus, zerolog, etc.)
    ```GO
	Debug(string)
	Info(string)
	Warn(string)
	Error(string)
    ```
