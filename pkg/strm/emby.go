package strm

import (
	"encoding/json"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type LibraryFolder struct {
	Name      string   `json:"Name"`
	ItemId    string   `json:"ItemId"`
	Locations []string `json:"Locations"`
}

var client = resty.New().SetRetryCount(3).SetRetryWaitTime(2 * time.Second)

func getHost() string {
	var host = conf.Conf.Emby.Host
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}
	return host
}

func getLibraryFoldersUrl() string {
	query := url.Values{}
	query.Set("X-Emby-Token", conf.Conf.Emby.APIKey)
	injectQuery, err := utils.InjectQuery(getHost()+"emby/Library/VirtualFolders", query)
	if err != nil {
		utils.Log.Errorf("Failed to inject query: %v", err)
	}
	return injectQuery
}

func getFreshLibraryUrl(itemId string) string {
	query := url.Values{}
	query.Set("X-Emby-Token", conf.Conf.Emby.APIKey)
	query.Set("Recursive", "true")
	injectQuery, err := utils.InjectQuery(getHost()+"emby/Items/"+itemId+"/Refresh", query)
	if err != nil {
		utils.Log.Errorf("Failed to inject query: %v", err)
	}
	return injectQuery
}

func getLibrary() []LibraryFolder {
	var result []LibraryFolder
	libraryFoldersUrl := getLibraryFoldersUrl()
	res, err := client.R().Get(libraryFoldersUrl)
	if err != nil {
		utils.Log.Fatalf("Failed to get library folders from URL [%s]: %v", libraryFoldersUrl, err)
	}

	if res.StatusCode() != 200 {
		utils.Log.Fatalf("Unexpected status code [%d] from URL [%s]", res.StatusCode(), libraryFoldersUrl)
	}
	if err := json.Unmarshal(res.Body(), &result); err != nil {
		utils.Log.Fatalf("Failed to parse response JSON: %v", err)
	}
	return result
}

func FreshLibrary(path string) []string {
	library := getLibrary()
	var freshLibrary []string
	path = filepath.Join(conf.Conf.StrmDir, path)
	for _, item := range library {
		for _, location := range item.Locations {
			if strings.HasPrefix(path, location) {
				libraryUrl := getFreshLibraryUrl(item.ItemId)
				_, err := client.R().Post(libraryUrl)
				if err != nil {
					utils.Log.Fatalf("Failed to get library folders from URL [%s]: %v", libraryUrl, err)
					continue
				}
				freshLibrary = append(freshLibrary, item.Name)
			}
		}
	}
	return freshLibrary
}
