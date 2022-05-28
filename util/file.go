package util

import (
	"os"
)

//判断文件或者目录是否存在
//如果返回的错误为nil,说明文件或文件夹存在
//如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
//如果返回的错误为其它类型,则不确定是否在存在
func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//创建空白文件
func CreateFile(filename string) bool {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}
