package main

import (
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

func FileExists(filename string) bool {
    _, err := os.Stat(filename)
    return !os.IsNotExist(err)
}

func EnsurePath(path string) error {
    // 使用 filepath.Clean 来处理路径，确保它是干净的
    cleanPath := filepath.Clean(path)

    // 检查路径是否已经存在
    _, err := os.Stat(cleanPath)
    if err == nil {
        // 目录已经存在，无需进一步处理
        return nil
    }

    if os.IsNotExist(err) {
        // 目录不存在，尝试创建它及其所有必要的父目录
        err := os.MkdirAll(cleanPath, os.ModePerm)
        if err != nil {
            return err
        }
        fmt.Printf("目录已创建：%s\n", cleanPath)
        return nil
    }

    // 发生了其他错误，返回错误信息
    return err
}

func RmFiles(paths ...string) {
    for _, path := range paths {
        if path == "" {
            continue
        }
        err := os.RemoveAll(path)
        if err != nil {
            err = errors.New(fmt.Sprint("RmFiles ", paths, err))
            fmt.Println("RmFiles ", err, paths)
        }
    }
    return
}

func GetFileSize(filePath string) (int64, error) {
    // 获取文件信息
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return 0, err
    }

    // 获取文件大小
    fileSize := fileInfo.Size()

    return fileSize, nil
}

func MvFile(sourceFile, targetFile string, overheadIfExist bool) (err error) {
    if !overheadIfExist {
        if _, err := os.Stat(targetFile); err == nil {
            err = errors.New("目标文件已经存在")
            return err
        }
    }
    var stdErr string
    _, stdErr, _, err = ExecCommand(fmt.Sprintf("mv %s %s", sourceFile, targetFile))
    if err == nil && stdErr != "" {
        err = errors.New(stdErr)
    }
    // err = os.Rename(sourceFile, targetFile) 不同设备之间只能mv 不能rename
    return
}

func EmptyDir(path string) (bool, error) {
    dir, err := os.Open(path)
    if err != nil {
        return false, err
    }
    defer dir.Close()

    entries, err := dir.Readdir(1)
    if err != nil {
        if err == io.EOF {
            return true, nil
        }
        return false, err
    }
    return len(entries) == 0, nil
}
