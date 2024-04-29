package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "os"
    "strings"
)

const (
    Usage = `
使用方法

./tcosTool  cosServer  uploadFile    localFile  cosFile  （cosFile 需要是完整带路径文件名，或者是以/结尾的路径）
./tcosTool  cosServer  uploadDir     localDir   cosDir
./tcosTool  cosServer  downloadFile  cosFile   localFile    
./tcosTool  cosServer  downloadDir   cosDir    localDir   

cosServer 可以是内置的cos桶 支持 "dcflow"、"ha3picw"、"ha3page"、"ha3inv"
          也可以是以 .yaml 结尾的桶完整配置文件名
          也可以是简化配置  bucket:ak:sk

eg 
./tcosTool  dcflow                            uploadFile   ./a.txt          tmp/a.txt
./tcosTool  mybucket.yaml                     downloadFile ./aDownload.txt  tmp/a.txt
./tcosTool  lzx-searchbs-ha3picw:AKxxx:xxxx   uploadDir    ./dir1           tmp/dir1



简化配置例子  lzx-searchbs-ha3picw:AKxxx:xxxx

完整配置例子 demo.yaml
bucket: lzx-searchbs-ha3picw
app_id: "1253445850"
secret_id: xxxx
secret_key: xxxx
region: ap-beijing
filesize_limit: 1000000000000
thumbnail: 1
directory: ""
is_need_proxy: false
cloud: tencent

`
)

func main() {
    if len(os.Args) < 5 {
        fmt.Println("参数错误")
        fmt.Println(Usage)
        return
    }

    // Extract command-line arguments
    cosServer := os.Args[1]
    operation := os.Args[2]
    localPath := os.Args[3]
    cosPath := os.Args[4]
    if strings.Contains(operation, "download") {
        localPath, cosPath = cosPath, localPath
    }

    var cos *DcTCos
    var err error

    // Determine the type of cosServer and create a DcTCos object accordingly
    if strings.HasSuffix(cosServer, ".yaml") {
        cos = LoadParamCosYaml(cosServer)
    } else if strings.Contains(cosServer, ":") {
        cos = LoadParamCosSimple(cosServer)
    } else {
        cos = TCosOriNameMap[cosServer]
    }

    if cos == nil {
        fmt.Println("Invalid cosServer argument.")
        return
    }

    ctx := gin.Context{}

    // Execute the operation based on the command-line arguments
    switch operation {
    case "uploadFile":
        err = cos.UploadFile(&ctx, localPath, cosPath, 3)
    case "uploadDir":
        err = cos.UploadLocalDir(&ctx, localPath, cosPath)
    case "downloadFile":
        err = cos.DownFile2local(&ctx, cosPath, localPath)
    case "downloadDir":
        err = cos.DownloadDir2Local(&ctx, cosPath, localPath)
    default:
        fmt.Println("Invalid operation. Please choose from uploadFile, uploadDir, downloadFile, or downloadDir.")
        return
    }

    if err != nil {
        fmt.Printf("Error during %s: %v\n", operation, err)
    } else {
        fmt.Printf("%s completed successfully.\n", strings.Title(operation))
    }
}
