package runner

import (
	"context"
	"fmt"
	"os"
	"path"
	"tbc/internal/terabox"

	"golang.org/x/sync/errgroup"
)

func Download(
	ctx context.Context,
	client *terabox.Client,
	remotePaths []string,
	destDir string,
	opts *terabox.DownloadOptions,
) error {
	listChan := make(chan terabox.DownloadFile, opts.MaxConcurrency)

	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(listChan)
		return listRecursive(egctx, client, remotePaths, destDir, listChan)
	})

	err := client.Download(ctx, nil, listChan, opts)
	if err != nil {
		return err
	}

	return eg.Wait()
}

func listRecursive(
	ctx context.Context,
	client *terabox.Client,
	remotePaths []string,
	destDir string,
	sender chan<- terabox.DownloadFile,
) error {
	files, dirs := List(ctx, client, remotePaths, terabox.OrderByName, false)

	if len(files.GetItems()) == 0 && len(dirs) == 0 {
		return fmt.Errorf("No valid file or directory specified")
	}

	for _, file := range files.GetItems() {
		select {
		case <-ctx.Done():
			return nil
		case sender <- terabox.DownloadFile{
			RemotePath: file.Path,
			LocalDir:   destDir,
		}:
		}
	}

	for _, dir := range dirs {
		baseDir := path.Join(destDir, path.Base(path.Clean(dir.OriginalArg)))
		if stat, err := os.Stat(baseDir); err != nil {
			if err := os.Mkdir(baseDir, os.ModePerm); err != nil {
				return fmt.Errorf("Failed to create directory")
			}
		} else if !stat.IsDir() {
			return fmt.Errorf("\"%s\" is not directory", dir.OriginalArg)
		}

		var innerDirs []string
		for _, item := range dir.GetItems() {
			if item.IsDir == 0 {
				select {
				case <-ctx.Done():
					return nil
				case sender <- terabox.DownloadFile{
					RemotePath: item.Path,
					LocalDir:   baseDir,
				}:
				}
			} else {
				innerDirs = append(innerDirs, item.Path)
			}
		}

		if len(innerDirs) > 0 {
			if err := listRecursive(ctx, client, innerDirs, baseDir, sender); err != nil {
				return err
			}
		}
	}
	return nil
}
