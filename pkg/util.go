package pkg

import (
	"io/fs"
	"os"
)

func ReadFileToString(scriptFilePath, fileName string) (string, error) {
	fsys := os.DirFS(scriptFilePath)
    b, err := fs.ReadFile(fsys, fileName)
    if err != nil {
        return "", err
    }
    
    return string(b), nil
}
