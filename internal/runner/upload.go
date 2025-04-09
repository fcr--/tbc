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

func Upload(
	ctx context.Context,
	client *terabox.Client,
	args []string,
	dest string,
	opts *terabox.UploadOptions,
) error {
	upFiles := []terabox.UploadFile{}
	dest = util.GetAbsPath(client.GetCwd(), dest)

	for _, file := range args {
		info, err := os.Stat(file)
		if err != nil {
			reason := strings.SplitN(err.Error(), ":", 2)
			if len(reason) > 1 {
				fmt.Fprintf(os.Stderr, "%s: %s\n", file, reason[1])
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}

		if info.IsDir() {
			files, err := util.FindFilesRecursive(info.Name())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			for _, dirfile := range files {
				upFiles = append(upFiles, terabox.UploadFile{
					LocalPath: dirfile,
					RemoteDir: path.Join(dest, path.Dir(dirfile)),
					FileName:  path.Base(dirfile),
				})
			}
		} else {
			upFiles = append(upFiles, terabox.UploadFile{
				LocalPath: file,
				RemoteDir: dest,
				FileName:  info.Name(),
			})
		}
	}

	if len(upFiles) == 0 {
		return fmt.Errorf("No files found to upload.")
	}
	return client.Upload(upFiles, opts)
}
