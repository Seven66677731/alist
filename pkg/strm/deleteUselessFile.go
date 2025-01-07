package strm

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

// getLocalFiles 递归获取目录下的所有文件
func getLocalFiles(path string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(filepath.Join(conf.Conf.StrmDir, path), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 跳过目录，只添加文件
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// 获取路径和文件名（去掉文件后缀）
func getPathWithFileNameWithoutExt(filePath string) string {
	dir := filepath.Dir(filePath)
	fileNameWithoutExt := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	return filepath.Join(dir, fileNameWithoutExt)
}

// DeleteExtraFiles 删除多余的文件
func DeleteExtraFiles(path string, alistFiles []string) []string {
	allFiles, _ := getLocalFiles(path)

	// 构建远程文件名集合（包含路径，忽略后缀）
	alistFileSet := make(map[string]struct{})
	for _, alistPath := range alistFiles {
		fullPath := filepath.Join(conf.Conf.StrmDir, alistPath)
		key := getPathWithFileNameWithoutExt(fullPath)
		alistFileSet[key] = struct{}{}
	}

	// 找出本地多余的文件
	var extraFiles []string
	for _, localFile := range allFiles {
		localFileKey := getPathWithFileNameWithoutExt(localFile)
		if _, exists := alistFileSet[localFileKey]; !exists {
			extraFiles = append(extraFiles, localFile)
		}
	}

	var resFiles []string
	// 删除文件
	for _, file := range extraFiles {
		err := os.Remove(file)
		if err != nil {
			utils.Log.Fatalf("Failed to delete file: %s, error: %v\n", file, err)
		} else {
			resFiles = append(resFiles, file)
			utils.Log.Infof("Deleted file: %s\n", file)
		}
	}
	return resFiles
}
