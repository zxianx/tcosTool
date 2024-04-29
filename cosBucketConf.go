package main

import "time"

type BucketConfig struct {
    Service       string `yaml:"service"`
    Bucket        string `yaml:"bucket"`
    AppID         string `yaml:"app_id"`
    SecretID      string `yaml:"secret_id"`
    SecretKey     string `yaml:"secret_key"`
    Region        string `yaml:"region"`
    FileSizeLimit int    `yaml:"filesize_limit"`
    Thumbnail     int    `yaml:"thumbnail"`
    Directory     string `yaml:"directory"`
    FilePrefix    string `yaml:"file_prefix"`
    Cloud         string `yaml:"cloud"`
    IsPublic      int    `yaml:"is_public"`
    CnameEnabled  bool   `yaml:"cnameEnabled"`

    MaxIdleConns        int
    MaxIdleConnsPerHost int
    MaxConnsPerHost     int
    IdleConnTimeout     time.Duration
    ConnectTimeout      time.Duration
}

var BaseConf BucketConfig = BucketConfig{
    Service:             "simple",
    Bucket:              "",
    AppID:               "1253445850",
    SecretID:            "",
    SecretKey:           "",
    Region:              "ap-beijing",
    FileSizeLimit:       10000000000000,
    Thumbnail:           1,
    Directory:           "",
    FilePrefix:          "",
    Cloud:               "tencent",
    IsPublic:            0,
    CnameEnabled:        false,
    MaxIdleConns:        0,
    MaxIdleConnsPerHost: 0,
    MaxConnsPerHost:     0,
    IdleConnTimeout:     0,
    ConnectTimeout:      0,
}

const innerCos = `
dc1flow:
  bucket: lzx-dcflow
  app_id: "1253445850"
  secret_id: x
  secret_key: x
  region: ap-beijing
  picture_region: picbj
  filesize_limit: 1048576
  thumbnail: 1
  directory: ""
  file_prefix: lzxtk_
  is_need_proxy: false
  cloud: tencent
demo2:
  

`
