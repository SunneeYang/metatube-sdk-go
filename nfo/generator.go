package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 设定目标根目录路径
	targetDirectory := "/Volumes/library-1/TODO"

	// 遍历目标目录下的所有文件
	_ = filepath.Walk(targetDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 过滤出视频文件，这里以.mp4为例
		if filepath.Ext(path) == ".mp4" {
			// 获取不带扩展名的文件名
			fileNameWithoutExt := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			// 构建NFO文件内容
			nfoContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
    <title>%s</title>
    <metatubeid>fc2hub</metatubeid>
</movie>`, fileNameWithoutExt)

			// 定义NFO文件路径
			nfoFilePath := fmt.Sprintf("%s.nfo", strings.TrimSuffix(path, filepath.Ext(path)))

			// 写入NFO文件
			if err := ioutil.WriteFile(nfoFilePath, []byte(nfoContent), 0644); err != nil {
				fmt.Println("Failed to write NFO file for", info.Name(), ":", err)
			} else {
				fmt.Println("NFO file created for:", info.Name())
			}
		}

		return nil
	})
}
