package terabox

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"tbc/internal/util"

	"github.com/tidwall/gjson"
)

type Item struct {
	Name     string `json:"server_filename"`
	IsDir    int    `json:"isdir"`
	Share    int    `json:"share"`
	Modified uint   `json:"local_mtime"`
	Created  uint   `json:"local_ctime"`
	Path     string `json:"path"`
	Size     uint   `json:"size"`
	FileId   uint64 `json:"fs_id"`
}

type OrderType string

const (
	OrderByName OrderType = "name"
	OrderByTime OrderType = "time"
	OrderBySize OrderType = "size"
)

type ListOptions struct {
	RemoteDir string
	OrderBy   OrderType
	Desc      bool
}

func (c *Client) List(opts ListOptions) ([]*Item, error) {
	remoteDir := util.GetAbsPath(c.cwd, opts.RemoteDir)

	var orderBy string
	switch opts.OrderBy {
	case OrderByName, OrderByTime, OrderBySize:
		orderBy = string(opts.OrderBy)
	default:
		orderBy = string(OrderByName)
	}
	var desc string
	if opts.Desc {
		desc = "1"
	} else {
		desc = "0"
	}

	var result []*Item
	hasMore := true
	page := 1
	for hasMore {
		body, err := util.GetResponse(c.client.R().
			SetQueryParams(map[string]string{
				"app_id":  appId,
				"jsToken": c.jsToken,
				"dir":     remoteDir,
				"order":   orderBy,
				"desc":    desc,
				"num":     "100",
				"page":    fmt.Sprintf("%d", page),
			}).
			Get("/api/list"))
		if err != nil {
			return nil, fmt.Errorf("ls: %s: Request failed: %w", opts.RemoteDir, err)
		}

		errNo := gjson.GetBytes(body, "errno").Int()
		switch errNo {
		case 0:
			break
		case -9:
			return nil, fmt.Errorf("ls: %s: No such directory", opts.RemoteDir)
		case -7:
			return nil, fmt.Errorf("ls: %s: Specified path is invalid", opts.RemoteDir)
		default:
			return nil, fmt.Errorf("ls: %s: Unexpected error (errno %d)", opts.RemoteDir, errNo)
		}
		list := gjson.GetBytes(body, "list")
		if !list.Exists() {
			return nil, fmt.Errorf("ls: %s: No list (errno %d)", opts.RemoteDir, errNo)
		}

		var items []*Item
		err = json.Unmarshal([]byte(list.Raw), &items)
		if err != nil {
			return nil, fmt.Errorf("ls: %s: Error parsing list: %w", opts.RemoteDir, err)
		}

		result = append(result, items...)
		page++
		if len(items) < 100 {
			break
		}
	}

	return result, nil
}

type SearchOptions struct {
	Keyword     string
	BaseDir     string
	NoRecursion bool
	OrderBy     OrderType
	Desc        bool
}

func (c *Client) Search(opts SearchOptions) ([]Item, error) {
	keyword := strings.TrimSpace(opts.Keyword)
	baseDir := util.GetAbsPath(c.cwd, opts.BaseDir)
	orderBy := string(OrderByName)
	switch opts.OrderBy {
	case OrderByName, OrderByTime, OrderBySize:
		orderBy = string(opts.OrderBy)
	}

	desc := "0"
	if opts.Desc {
		desc = "1"
	}

	params := map[string]string{
		"app_id":    appId,
		"jsToken":   c.jsToken,
		"order":     orderBy,
		"desc":      desc,
		"dir":       baseDir,
		"num":       "100",
		"page":      "1",
		"recursion": "1",
		"key":       keyword,
	}

	if opts.NoRecursion {
		delete(params, "recursion")
	}

	var result []Item
	hasMore := true
	page := 1
	for hasMore {
		params["page"] = fmt.Sprintf("%d", page)
		body, err := util.GetResponse(c.client.R().
			SetQueryParams(params).
			Get("/api/search"))
		if err != nil {
			return nil, fmt.Errorf("find: Request failed: %w", err)
		}

		errNo := gjson.GetBytes(body, "errno").Int()
		switch errNo {
		case 0:
			break
		default:
			return nil, fmt.Errorf("find: Unexpected error (errno %d)", errNo)
		}
		list := gjson.GetBytes(body, "list")
		if !list.Exists() {
			return nil, fmt.Errorf("find: No list (errno %d)", errNo)
		}

		var items []Item
		err = json.Unmarshal([]byte(list.Raw), &items)
		if err != nil {
			return nil, fmt.Errorf("find: Error parsing list: %w", err)
		}

		result = append(result, items...)
		hasMore = list.Get("has_more").Int() != 0
		page++
	}

	return result, nil
}

func (c *Client) MakeDir(directory string) (bool, error) {
	remoteDir := util.GetAbsPath(c.cwd, directory)
	body, err := util.GetResponse(c.client.R().
		SetQueryParams(map[string]string{
			"a":        "commit",
			"app_id":   appId,
			"jsToken":  c.jsToken,
			"bdstoken": c.bdsToken,
		}).
		SetFormData(map[string]string{
			"path":       remoteDir,
			"isdir":      "1",
			"block_list": "[]",
		}).
		Post("/api/create"))
	if err != nil {
		return false, fmt.Errorf("mkdir: %s: Request failed: %w", directory, err)
	}

	errNo := gjson.GetBytes(body, "errno").Int()
	if errNo != 0 {
		return false, fmt.Errorf("mkdir: Unexpected error (errno %d)", errNo)
	}

	return true, nil
}

func (c *Client) Remove(remotePaths []string) (bool, error) {
	for i := range remotePaths {
		remotePaths[i] = util.GetAbsPath(c.cwd, remotePaths[i])
		if remotePaths[i] == "/" {
			return false, fmt.Errorf("rm: Cannot remove root directory")
		}
	}
	filelist, _ := json.Marshal(remotePaths)
	body, err := util.GetResponse(c.client.R().
		SetQueryParams(map[string]string{
			"async":    "0",
			"onnest":   "fail",
			"opera":    "delete",
			"app_id":   appId,
			"jsToken":  c.jsToken,
			"bdstoken": c.bdsToken,
		}).
		SetFormData(map[string]string{
			"filelist": string(filelist),
		}).
		Post("/api/filemanager"))
	if err != nil {
		return false, fmt.Errorf("rm: Request failed: %w", err)
	}

	errNo := gjson.GetBytes(body, "errno").Int()
	switch errNo {
	case 0:
		break
	case 2:
		return false, fmt.Errorf("rm: No such file or directory")
	default:
		return false, fmt.Errorf("rm: Unexpected error (errno %d)", errNo)
	}
	return true, nil
}

type moveFileFormData struct {
	SourcePath  string `json:"path"`
	Destination string `json:"dest"`
	OnDuplicate string `json:"ondup,omitempty"`
	NewName     string `json:"newname"`
}

type FileManageOperation string

const (
	OpMove FileManageOperation = "move"
	OpCopy FileManageOperation = "copy"
)

type OnDuplicateOption string

const (
	OpOverwrite OnDuplicateOption = "overwrite"
	OpNewCopy   OnDuplicateOption = "newcopy"
	OpNone      OnDuplicateOption = ""
)

func (c *Client) fileManage(op FileManageOperation, sources *[]string, destination string, ondup OnDuplicateOption) (bool, error) {
	var opera string
	switch op {
	case OpMove, OpCopy:
		opera = string(op)
	default:
		return false, fmt.Errorf("Operation not supported")
	}

	var dup string
	switch ondup {
	case OpOverwrite, OpNewCopy:
		dup = string(ondup)
	default:
		dup = ""
	}

	if len(*sources) == 0 {
		return false, fmt.Errorf("No file specified")
	}

	destDir := util.GetAbsPath(c.cwd, destination)
	if !strings.HasSuffix(destDir, "/") && strings.HasSuffix(destination, "/") {
		destDir += "/"
	}

	var formData []moveFileFormData

	// Check destination directory is available
	if _, err := c.List(ListOptions{
		RemoteDir: destDir,
	}); err != nil {
		if len(*sources) > 1 {
			return false, fmt.Errorf("Destination directory not found")
		}

		// Check parent destination  directory is available
		destDir = path.Dir(destDir)
		if _, err := c.List(ListOptions{
			RemoteDir: destDir,
		}); err != nil {
			return false, fmt.Errorf("Destination directory not found")
		}

		srcPath := (*sources)[0]
		if !strings.HasSuffix(srcPath, "/") {
			srcPath = path.Join(c.cwd, srcPath)
		}

		formData = append(formData, moveFileFormData{
			SourcePath:  srcPath,
			Destination: destDir,
			OnDuplicate: dup,
			NewName:     path.Base(destination),
		})
	}

	if len(formData) == 0 {
		for _, srcPath := range *sources {
			srcPath = util.GetAbsPath(c.cwd, srcPath)
			formData = append(formData, moveFileFormData{
				SourcePath:  srcPath,
				Destination: destDir,
				OnDuplicate: dup,
				NewName:     path.Base(srcPath),
			})
		}
	}

	filelist, _ := json.Marshal(formData)
	async := "0"
	if ondup == OpOverwrite {
		async = "2"
	}
	body, err := util.GetResponse(c.client.R().
		SetQueryParams(map[string]string{
			"async":    async,
			"onnest":   "fail",
			"opera":    opera,
			"app_id":   appId,
			"jsToken":  c.jsToken,
			"bdstoken": c.bdsToken,
		}).
		SetFormData(map[string]string{
			"filelist": string(filelist),
		}).
		Post("/api/filemanager"))
	if err != nil {
		return false, fmt.Errorf("Request failed: %w", err)
	}

	errNo := gjson.GetBytes(body, "errno").Int()
	switch errNo {
	case 0:
		break
	case 12:
		return false, fmt.Errorf("Some files moving failed")
	default:
		return false, fmt.Errorf("Unexpected error (errno %d)", errNo)
	}
	return true, nil
}

func (c *Client) Move(sources []string, destination string, ondup OnDuplicateOption) (bool, error) {
	result, err := c.fileManage(OpMove, &sources, destination, ondup)
	if err != nil {
		return false, fmt.Errorf("mv: %w", err)
	}
	return result, nil
}

func (c *Client) Copy(sources []string, destination string, ondup OnDuplicateOption) (bool, error) {
	result, err := c.fileManage(OpCopy, &sources, destination, ondup)
	if err != nil {
		return false, fmt.Errorf("cp: %w", err)
	}
	return result, nil
}

func (c *Client) ChDir(dir string) error {
	dir = util.GetAbsPath(c.cwd, dir)
	if _, err := c.List(ListOptions{RemoteDir: dir}); err != nil {
		return fmt.Errorf("cd: %s: No such directory", dir)
	}
	c.cwd = dir
	return nil
}

func (c *Client) GetCwd() string {
	return c.cwd
}
