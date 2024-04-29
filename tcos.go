package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/gin-gonic/gin"
    errors2 "github.com/pkg/errors"
    tcos "github.com/tencentyun/cos-go-sdk-v5"
    "io"
    "net/http"
    "net/url"
    "os"
    "path"
    "path/filepath"
    "runtime"
    "strconv"
    "strings"
    "time"
)

// 公司提供的上传cos会自动加 jpg后缀，虽然可以指定后缀，但是对于无后缀的文件名无法处理

type DcTCos struct {
    cli *tcos.Client
}

func NewTcosBucket(bucket, appid, region, ak, sk string) (res *DcTCos, err error) {
    c := &http.Client{
        Transport: &tcos.AuthorizationTransport{
            SecretID:  ak,
            SecretKey: sk,
            Expire:    time.Hour,
        },
    }
    b, err := getBucketUrl(bucket, appid, region)
    if err != nil {
        return nil, err
    }
    basicUrl := &tcos.BaseURL{
        BucketURL: b,
    }
    client := tcos.NewClient(basicUrl, c)
    res = &DcTCos{
        cli: client,
    }
    return
}

func getBucketUrl(bucket, appID, region string) (*url.URL, error) {
    bucketUrl := fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", bucket, appID, region)
    return url.Parse(bucketUrl)
}

func (dcTCos *DcTCos) UploadLocalDirDebug(ctx *gin.Context, srcDir string, targetDir string) (err error) {
    err = filepath.Walk(srcDir, func(localPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        var emptyDir bool
        if info.IsDir() {
            emptyDir, err = EmptyDir(localPath)
            if err != nil {
                return err
            }
            if !emptyDir {
                return nil
            }
        }

        // 相对路径，去除 srcDir 的前缀
        relativePath, err := filepath.Rel(srcDir, localPath)
        if err != nil {
            return err
        }

        // 构建目标 COS 路径
        targetPath := path.Join(targetDir, relativePath)

        if emptyDir {
            // 空文件夹上传
            targetPath += "/"
            //   fmt.Println("emptyDir", localPath, info.Name(), targetPath, localPath)
            var resp *tcos.Response
            resp, err = dcTCos.cli.Object.Put(ctx, targetPath, nil, nil)
            _ = closeTcosResp(resp)
        } else {
            err = dcTCos.UploadFile(ctx, localPath, targetPath, 4)
        }
        if err != nil {
            return err
        }
        return nil
    })
    return
}

func (dcTCos *DcTCos) UploadLocalDir(ctx *gin.Context, srcDir string, targetDir string) error {
    if targetDir != "" && targetDir[0] != '/' {
        targetDir = "/" + targetDir
    }
    return dcTCos.UploadLocalDirDebug(ctx, srcDir, targetDir)
}

// GetSubFileDir 获取指定目录下的文件和子目录列表
func (dcTCos *DcTCos) GetSubFileDir(ctx *gin.Context, parentDir string, resJoinParentDir bool) (files []string, dirs []string, err error) {
    if parentDir == "" {
        parentDir = "/"
    }
    if parentDir[0] == '/' {
        parentDir = parentDir[1:]
    }
    if parentDir[len(parentDir)-1] != '/' {
        parentDir += "/"
    }

    listOptions := &tcos.BucketGetOptions{
        Prefix:    parentDir,
        Delimiter: "/",
        MaxKeys:   1000, // 调整最大返回结果数量
    }

    // 发起请求
    v, resp, err := dcTCos.cli.Bucket.Get(context.Background(), listOptions)
    if err != nil {
        return nil, nil, err
    }
    _ = closeTcosResp(resp)
    for _, object := range v.Contents {
        if resJoinParentDir {
            files = append(files, object.Key)
        } else {
            // 截取文件名
            filename := object.Key
            files = append(files, filename[len(parentDir):])
        }
    }

    for _, prefix := range v.CommonPrefixes {
        prefix = prefix[:len(prefix)-1]
        if resJoinParentDir {
            dirs = append(dirs, prefix)
        } else {
            // 截取子目录名
            subdir := prefix
            dirs = append(dirs, subdir[len(parentDir):])
        }
    }

    return files, dirs, nil
}

func (dcTCos *DcTCos) delDir(ctx *gin.Context, dir string) (err error) {
    files, dirs, err := dcTCos.GetSubFileDir(ctx, dir, true)
    err = dcTCos.DelFiles(ctx, files)
    if err != nil {
        return err
    }
    for _, s := range dirs {
        err = dcTCos.DelDir(ctx, s)
        if err != nil {
            return err
        }
    }
    return
}

func (dcTCos *DcTCos) DelDir(ctx *gin.Context, dir string) (err error) {
    if strings.Trim(dir, "/.") == "" {
        // 待完善
        err = errors.New("forbid")
        return
    }
    if dir[len(dir)-1] != '/' {
        dir += "/"
    }
    return dcTCos.delDir(ctx, dir)
}

func (dcTCos *DcTCos) DelFile(ctx *gin.Context, dir string) (err error) {
    resp, err := dcTCos.cli.Object.Delete(ctx, dir, nil)
    _ = closeTcosResp(resp)
    return
}

func (dcTCos *DcTCos) DelFiles(ctx *gin.Context, dir []string) (err error) {
    // 官方 DeleteMulti 没看懂参数
    for _, s := range dir {
        err = dcTCos.DelFile(ctx, s)
        if err != nil {
            return err
        }
    }
    return
}

func closeTcosResp(resp *tcos.Response) (err error) {
    if resp == nil || resp.Body == nil {
        return
    }
    _, err = io.Copy(io.Discard, resp.Body)
    if err != nil {
        return err
    }
    err = resp.Body.Close()
    return
}

func (dcTCos *DcTCos) UploadFile(ctx *gin.Context, localPath, targetPath string, retry int) (err error) {
    l := len(targetPath) - 1
    if targetPath[l] == '/' || targetPath[l] == '\\' {
        targetPath += filepath.Base(localPath)
    }

    defer fmt.Println("debugUploadToCos", targetPath, localPath, err)
    //file, err := os.Open(localPath)
    //if err != nil {
    //    return err
    //}
    //defer file.Close()
    retrySleep := 1 * time.Millisecond
    var uploadErr error
    var resp *tcos.Response
    retryBack := retry
    for retry+1 > 0 {
        _, resp, uploadErr = dcTCos.cli.Object.Upload(ctx, targetPath, localPath, nil)
        //   resp, uploadErr = dcTCos.cli.Object.Put(ctx, targetPath, file, nil)
        _ = closeTcosResp(resp)
        if uploadErr == nil {
            return
        }
        time.Sleep(retrySleep)
        retrySleep *= 2
        retry--
    }
    if uploadErr != nil {
        err = errors.New(fmt.Sprintf("upload Err retry[%d],err[%v]", retryBack, uploadErr))
    }
    return
}

func (dcTCos *DcTCos) downFile2local(cosKey, localDirName string) (err error) {
    var resp *tcos.Response
    resp, err = dcTCos.cli.Object.Get(context.TODO(), cosKey, nil)
    if err != nil {
        return
    }
    defer resp.Body.Close()
    localFile, err := os.Create(localDirName)
    if err != nil {
        return
    }
    defer localFile.Close()
    // 将对象内容写入本地文件
    _, err = io.Copy(localFile, resp.Body)
    if err != nil {
        return
    }
    return
}

func (dcTCos *DcTCos) DownDir2local(ctx *gin.Context, cosKey, localDirName string) (err error) {
    retry := 1
    defer func() {
        fmt.Println("DownFile2local", cosKey, localDirName, err)
    }()
    if cosKey == "" || localDirName == "" {
        return errors.New("empty cosKey or localDirName")
    }
    //if cosKey[0] != '/' {
    //    cosKey="/"+cosKey
    //}
    localDir := filepath.Dir(localDirName)
    err = os.MkdirAll(localDir, os.ModePerm)
    if err != nil {
        return WithErrMessageAndStack(err, "mkdir"+localDir)
    }
    retrySleep := 1 * time.Millisecond
    var downLoadErr error
    retryBack := retry
    var trimLeft bool
    for retry+1 > 0 {
        downLoadErr = dcTCos.downFile2local(cosKey, localDirName)
        if downLoadErr == nil {
            return
        }
        if !trimLeft && cosKey[0] == '/' && strings.Contains(downLoadErr.Error(), "404 NoSuchKey") {
            cosKey = cosKey[1:]
            continue
        }

        time.Sleep(retrySleep)
        retrySleep *= 2
        retry--
    }
    if downLoadErr != nil {
        err = errors.New(fmt.Sprintf("downLoad Err retry[%d],err[%v]", retryBack, downLoadErr))
    }
    return
}

func (dcTCos *DcTCos) DownFile2local(ctx *gin.Context, cosKey, localDirName string) (err error) {
    retry := 1
    defer func() {
        fmt.Println("DownFile2local", cosKey, localDirName, err)
    }()
    if cosKey == "" || localDirName == "" {
        return errors.New("empty cosKey or localDirName")
    }
    //if cosKey[0] != '/' {
    //    cosKey="/"+cosKey
    //}
    localDir := filepath.Dir(localDirName)
    err = os.MkdirAll(localDir, os.ModePerm)
    if err != nil {
        return WithErrMessageAndStack(err, "mkdir"+localDir)
    }
    retrySleep := 1 * time.Millisecond
    var downLoadErr error
    retryBack := retry
    var trimLeft bool
    for retry+1 > 0 {
        downLoadErr = dcTCos.downFile2local(cosKey, localDirName)
        if downLoadErr == nil {
            return
        }
        if !trimLeft && cosKey[0] == '/' && strings.Contains(downLoadErr.Error(), "404 NoSuchKey") {
            cosKey = cosKey[1:]
            continue
        }

        time.Sleep(retrySleep)
        retrySleep *= 2
        retry--
    }
    if downLoadErr != nil {
        err = errors.New(fmt.Sprintf("downLoad Err retry[%d],err[%v]", retryBack, downLoadErr))
    }
    return
}

func (dcTCos *DcTCos) DownloadDir2Local(ctx *gin.Context, cosDir, localDir string) (err error) {
    defer func() {
        fmt.Println("DownloadDir2Local", cosDir, localDir, err)
    }()

    if cosDir == "" || localDir == "" {
        return errors.New("empty cosDir or localDir")
    }

    // Ensure the local directory exists
    err = os.MkdirAll(localDir, os.ModePerm)
    if err != nil {
        return WithErrMessageAndStack(err, "mkdir "+localDir)
    }

    // List files and subdirectories in the COS directory
    files, dirs, err := dcTCos.GetSubFileDir(ctx, cosDir, false)
    if err != nil {
        return WithErrMessageAndStack(err, "failed to list files and directories in COS directory")
    }

    // Download files in the COS directory
    for _, file := range files {
        cosKey := filepath.Join(cosDir, file)
        localFilePath := filepath.Join(localDir, file)

        err = dcTCos.DownFile2local(ctx, cosKey, localFilePath)
        if err != nil {
            err = WithErrMessageAndStack(err, fmt.Sprintf("dcTCos DownFile2local[%s]=>[%s]", cosKey, localFilePath))
            return
        }
    }

    // Recursively download subdirectories
    for _, subdir := range dirs {
        cosSubDir := filepath.Join(cosDir, subdir) + "/"
        localSubDir := filepath.Join(localDir, subdir)

        err = dcTCos.DownloadDir2Local(ctx, cosSubDir, localSubDir)
        if err != nil {
            err = WithErrMessageAndStack(err, "dcTCos.DownloadDir2Local.i")
            return
        }
    }

    return nil
}

func WithErrMessageAndStack(err error, message string) error {
    return errors2.WithMessage(err, "==> "+printCallerNameAndLine()+message)
}

func printCallerNameAndLine() string {
    pc, _, line, _ := runtime.Caller(2)
    return runtime.FuncForPC(pc).Name() + "@" + strconv.Itoa(line) + " "
}
