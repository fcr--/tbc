package runner

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"tbc/internal/terabox"
	"tbc/internal/util"
)

type FilesResults struct {
	DisplayPath []string
	Items       []*terabox.Item
}

type DirResults struct {
	OriginalArg string
	Path        string
	Items       []*terabox.Item
}

type ListResults interface {
	GetItems() []*terabox.Item
	GetDisplayPath(index int) string
}

func (r *FilesResults) GetItems() []*terabox.Item {
	return r.Items
}

func (r *FilesResults) Len() int {
	return len(r.DisplayPath)
}

func (r *FilesResults) Add(displayPath string, item *terabox.Item) {
	r.DisplayPath = append(r.DisplayPath, displayPath)
	r.Items = append(r.Items, item)
}

func (r *FilesResults) Join(source *FilesResults) {
	r.DisplayPath = append(r.DisplayPath, source.DisplayPath...)
	r.Items = append(r.Items, source.Items...)
}

func (r *FilesResults) GetDisplayPath(index int) string {
	if 0 <= index && index < len(r.DisplayPath) {
		return r.DisplayPath[index]
	}
	return ""
}

func (r *DirResults) GetItems() []*terabox.Item {
	return r.Items
}

func (r *DirResults) GetDisplayPath(index int) string {
	return r.Items[index].Name
}

func List(
	ctx context.Context,
	client *terabox.Client,
	args []string,
	sort terabox.OrderType,
	rev bool,
) (*FilesResults, []*DirResults) {
	if len(args) == 0 {
		args = []string{"."}
	}

	fetched := make(map[string][]*terabox.Item)
	items := make(map[string]*terabox.Item)
	files := &FilesResults{}
	dirs := []*DirResults{}

	for _, arg := range args {
		absPath := util.GetAbsPath(client.GetCwd(), arg)
		target := path.Dir(absPath)
		if _, exists := fetched[target]; !exists {
			res, err := client.List(terabox.ListOptions{
				RemoteDir: target,
				OrderBy:   sort,
				Desc:      rev,
			})
			if err == nil {
				fetched[target] = res
				for _, item := range res {
					items[item.Path] = item
				}
			}
		}

		if _, exists := fetched[target]; !exists {
			fmt.Fprintf(os.Stderr, `"%s"`+": No such file or directory\n", arg)
			continue
		}

		// check wildcard
		if strings.ContainsAny(path.Base(absPath), "*?[") {
			matched := &FilesResults{}
			for _, item := range fetched[target] {
				if ok, _ := path.Match(path.Base(absPath), item.Name); ok {
					matched.Add(getDisplayPath(arg, item.Path), item)
				}
			}

			if matched.Len() == 0 {
				fmt.Fprintf(os.Stderr, `"%s"`+": No such file or directory\n", arg)
				continue
			}
			files.Join(matched)
		} else if absPath == "/" {
			dirs = append(dirs, &DirResults{
				OriginalArg: arg,
				Path:        absPath,
				Items:       fetched[absPath],
			})
		} else if item, exists := items[absPath]; !exists {
			fmt.Fprintf(os.Stderr, `"%s"`+": No such file or directory\n", arg)
		} else {
			if item.IsDir != 0 {
				res, err := client.List(terabox.ListOptions{
					RemoteDir: item.Path,
					OrderBy:   sort,
					Desc:      rev,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, `"%s"`+": Fetch failed\n", arg)
					continue
				}
				dirs = append(dirs, &DirResults{
					OriginalArg: arg,
					Path:        absPath,
					Items:       res,
				})
			} else {
				files.Add(getDisplayPath(arg, item.Path), item)
			}
		}
	}

	return files, dirs
}

func getDisplayPath(originalPath, remotePath string) string {
	cleanPath := path.Clean(originalPath)
	if strings.HasPrefix(cleanPath, "/") {
		return remotePath
	}
	return path.Join(path.Dir(cleanPath), path.Base(remotePath))
}
