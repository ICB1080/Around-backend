package main

import (
    "context"
    "fmt"
    "io"

    "cloud.google.com/go/storage"
)

const (
    BUCKET_NAME = "around-web-bucket"
)

func saveToGCS(r io.Reader, objectName string) (string, error) {
    ctx := context.Background()

    client, err := storage.NewClient(ctx)
    if err != nil {
        return "", err
    }

    object := client.Bucket(BUCKET_NAME).Object(objectName)
    wc := object.NewWriter(ctx)
	// if里面inline的err只在if范围内使用，所以用冒号申明
    if _, err := io.Copy(wc, r); err != nil {
        return "", err
    }

    if err := wc.Close(); err != nil {
        return "", err
    }

	// 权限
    if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
        return "", err
    }

	// 读取文件属性
	// 冒号前面有一个未经定义的变量就可以两个一起用冒号了。
    attrs, err := object.Attrs(ctx)
    if err != nil {
        return "", err
    }

	// MediaLink文件所对应链接
    fmt.Printf("Image is saved to GCS: %s\n", attrs.MediaLink)
    return attrs.MediaLink, nil
}