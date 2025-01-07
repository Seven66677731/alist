package strm

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/conf"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	videoSuffix = map[string]struct{}{"mp4": {}, "mkv": {}, "flv": {}, "avi": {}, "wmv": {}, "ts": {}, "rmvb": {}, "webm": {}}
)

type WriteResult struct {
	Path    string
	Success bool
	Error   error
}

// getLocalFilePath 获取本地文件路径
func getLocalFilePath(filePath string) string {
	fileNameWithoutExt := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	if strings.HasPrefix(fileNameWithoutExt, "/") {
		return filepath.Join(conf.Conf.StrmDir, fileNameWithoutExt+".strm")
	}
	return filepath.Join(conf.Conf.StrmDir, fileNameWithoutExt+".strm")
}

// getFileContent 获取文件内容
func getFileContent(filePath string) string {
	encodedPath := url.PathEscape(filePath)
	return strings.TrimRight(conf.Conf.SiteURL, "/") + "/d" + encodedPath
}

// WriteFile 写入单个文件并返回写入结果
func WriteFile(filePath string) WriteResult {
	result := WriteResult{}

	// 跳过非视频文件
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	if _, ok := videoSuffix[ext]; !ok {
		result.Success = false
		result.Error = fmt.Errorf("文件类型不支持: %s", ext)
		return result
	}

	// 拼接目标文件路径
	localFilePath := getLocalFilePath(filePath)
	result.Path = localFilePath
	// 确保目录存在
	localDir := filepath.Dir(localFilePath)
	if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
		result.Success = false
		result.Error = fmt.Errorf("创建目录失败: %v", err)
		return result
	}

	// 写入文件内容
	content := getFileContent(filePath)
	if err := os.WriteFile(localFilePath, []byte(content), 0644); err != nil {
		result.Success = false
		result.Error = fmt.Errorf("写入文件失败: %v", err)
	} else {
		result.Success = true
	}
	return result
}

// WriteFiles 批量写入文件，并返回成功和失败的结果
func WriteFiles(filePaths []string) map[string]interface{} {
	var wg sync.WaitGroup
	results := make(chan WriteResult, len(filePaths))
	for _, filePath := range filePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			results <- WriteFile(path)
		}(filePath)
	}
	wg.Wait()
	close(results)

	// 处理结果
	var successPaths, failedPaths []string
	for result := range results {
		if result.Success {
			successPaths = append(successPaths, result.Path)
		} else {
			failedPaths = append(failedPaths, result.Path+"\n"+result.Error.Error())
		}
	}
	resMap := make(map[string]interface{})
	resMap["successPaths"] = successPaths
	resMap["failedPaths"] = failedPaths
	return resMap
}
