# params

Package for safe saving and holding parameters in file. Parameterss reading value from file when Get() and writing value into file when Set(), i.e. file not blocked all time.


File of parameters will be readed or created new if not exists. When reading existing file - will be matching it Data Types.

</br>

1. Init object for use it in scripts

    ```GO
    New(fName string, t DataType, pass string) (*ParamsObj, error)
    ```
    * fName - file name (in current work directory) or path to file;
    * t - type storages;
    * pass - password when data is stored encrypted.


2. Writing value to file, returns nil if the process was completed successfully

    ```GO
    SetValue(n string, val string) (error)
    ```
    * n - parameter name, the name starting with letter, can consist of letters, numbers, and symbols "-", "_";
    * val - writing value;


3. Reading value from file, returns value and nil if the process was completed successfully

    ```GO
    GetValue(n string) (val string, err error)
    ```
    * n - parameter name, the name starting with letter, can consist of letters, numbers, and symbols "-", "_";
    * val - returning value;
    

4. Delete value from file, returns nil if the process was completed successfully

    ```GO
    DeleteValue(n string) (err error)
    ```
    * n - parameter name;
    
