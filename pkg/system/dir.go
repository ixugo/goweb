package system

import (
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// 获取可执行文件绝对路径
var GetCWD = Executable()

func Executable() func() string {
	var once sync.Once
	var dir string
	return func() string {
		once.Do(func() {
			bin, _ := os.Executable()
			dir = filepath.Dir(bin)
		})
		return dir
	}
}

// GetDirSize 获取目录大小，单位 Bit
func GetDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func CleanOldFiles(path string, count int) error {
	var files []os.FileInfo
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, info)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 按照文件的修改时间升序排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	// 找到最旧的文件
	if len(files) > count {
		files = files[:count]
	}

	// 删除文件
	for _, file := range files {
		filePath := filepath.Join(path, file.Name())
		if err := os.Remove(filePath); err != nil {
			slog.Error("文件删除失败", "err", err)
		} else {
			slog.Info("删除旧文件", "path", filePath)
		}
	}
	return nil
}

// RemoveEmptyDirs 删除空目录
func RemoveEmptyDirs(rootDir string) error {
	// 遍历目录树，按反向顺序处理目录
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}
		// 检查目录是否为空
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			// 删除空目录
			if err := os.Remove(path); err != nil {
				slog.Error("删除空目录出错", "err", err)
			}
		}
		return nil
	})
}
