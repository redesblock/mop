package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/redesblock/mop/core/api"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (c *command) initUploadCmd() error {
	cmd := &cobra.Command{
		Use:   "upload id file",
		Short: "upload file or a collection of files.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var buf bytes.Buffer
			trawriter := tar.NewWriter(&buf)
			filesource := args[1]
			sfileInfo, err := os.Stat(filesource)
			if err != nil {
				return err
			}
			if !sfileInfo.IsDir() {
				tarFile(filesource, sfileInfo, trawriter)
			} else {
				tarFolder(filesource, trawriter)
			}
			if err := trawriter.Close(); err != nil {
				return err
			}

			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:1633/mop"), bytes.NewReader(buf.Bytes()))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/x-tar")
			//req.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
			req.Header.Set(api.FlockCollectionHeader, "true")
			req.Header.Set(api.FlockIndexDocumentHeader, sfileInfo.Name())
			req.Header.Set(api.FlockPostageBatchIdHeader, args[0])
			response, err := client.Do(req)
			if err != nil {
				return err
			}

			stdout := os.Stdout
			_, err = io.Copy(stdout, response.Body)

			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.config.BindPFlags(cmd.Flags())
		},
	}

	c.setAllFlags(cmd)
	c.root.AddCommand(cmd)

	return nil
}

func tarFile(filesource string, info os.FileInfo, tarwriter *tar.Writer) error {
	// 打开文件
	afile, err := os.Open(filesource)
	if err != nil {
		return err
	}
	// 关闭文件句柄
	defer afile.Close()
	// 头文件信息
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filesource
	// 写入文件
	err = tarwriter.WriteHeader(header)
	if err != nil {
		return err
	}

	if _, err = io.Copy(tarwriter, afile); err != nil {
		return err
	}
	return nil
}

func tarFolder(filesource string, trawriter *tar.Writer) error {
	return filepath.Walk(filesource, func(targetpath string, info fs.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			if filesource == targetpath {
				return nil
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = filepath.Join(filesource, strings.TrimPrefix(targetpath, filesource))
			if err = trawriter.WriteHeader(header); err != nil {
				return err
			}
			return tarFolder(targetpath, trawriter)
		}
		return tarFile(targetpath, info, trawriter)
	})
}
