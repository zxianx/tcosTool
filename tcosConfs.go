package main

import (
    "fmt"
    "gopkg.in/yaml.v2"
    "os"
    "strings"
)

var TCosOriNameMap = map[string]*DcTCos{"": nil}
var TCosOriNameConfMap = map[string]BucketConfig{}

func init() {
    var err error
    err = yaml.Unmarshal([]byte(innerCos), &TCosOriNameConfMap)
    if err != nil {
        panic("load innerCosFile Err yaml.Unmarshal " + err.Error())
        return
    }

    for cosMame, config := range TCosOriNameConfMap {
        cos, err := NewTcosBucket(config.Bucket, config.AppID, config.Region, config.SecretID, config.SecretKey)
        if err != nil {
            panic(fmt.Sprintf("load innerCosFile bucker[%s]err[%v]", cosMame, err))
            return
        }
        TCosOriNameMap[cosMame] = cos
    }

}

func LoadParamCosYaml(path string) (dctcos *DcTCos) {
    conf := BucketConfig{}
    yamlFile, err := os.ReadFile(path)
    if err != nil {
        panic("load LoadParamCosYaml Err readFile " + err.Error())
        return
    }
    err = yaml.Unmarshal(yamlFile, &conf)
    if err != nil {
        panic("load LoadParamCosYaml Err yaml.Unmarshal " + err.Error())
        return
    }
    dctcos, err = NewTcosBucket(conf.Bucket, conf.AppID, conf.Region, conf.SecretID, conf.SecretKey)
    if err != nil {
        panic(fmt.Sprintf("load LoadParamCosYaml err %v", err))
        return
    }
    return
}

func LoadParamCosSimple(confSimple string) (dctcos *DcTCos) {
    conf := BaseConf
    confSimpleArr := strings.Split(confSimple, ":")
    if len(confSimpleArr) != 3 {
        panic(fmt.Sprintf("LoadParamCosSimple  illegal"))
    }
    conf.Bucket = confSimpleArr[0]
    conf.SecretID = confSimpleArr[1]
    conf.SecretKey = confSimpleArr[2]
    var err error
    dctcos, err = NewTcosBucket(conf.Bucket, conf.AppID, conf.Region, conf.SecretID, conf.SecretKey)
    if err != nil {
        panic(fmt.Sprintf("load LoadParamCosSimple err %v", err))
        return
    }
    return
}
