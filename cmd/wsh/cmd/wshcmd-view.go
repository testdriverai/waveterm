// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/waveterm/pkg/waveobj"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
)

var viewMagnified bool

var viewCmd = &cobra.Command{
	Use:     "view {file|directory|URL}",
	Short:   "preview/edit a file or directory",
	Args:    cobra.ExactArgs(1),
	RunE:    viewRun,
	PreRunE: preRunSetupRpcClient,
}

var editCmd = &cobra.Command{
	Use:     "edit {file}",
	Short:   "edit a file",
	Args:    cobra.ExactArgs(1),
	RunE:    viewRun,
	PreRunE: preRunSetupRpcClient,
}

func init() {
	viewCmd.Flags().BoolVarP(&viewMagnified, "magnified", "m", false, "open view in magnified mode")
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(editCmd)
}

func viewRun(cmd *cobra.Command, args []string) (rtnErr error) {
	defer func() {
		sendActivity("view", rtnErr == nil)
	}()
	fileArg := args[0]
	conn := RpcContext.Conn
	var wshCmd *wshrpc.CommandCreateBlockData
	if strings.HasPrefix(fileArg, "http://") || strings.HasPrefix(fileArg, "https://") {
		wshCmd = &wshrpc.CommandCreateBlockData{
			BlockDef: &waveobj.BlockDef{
				Meta: map[string]any{
					waveobj.MetaKey_View: "web",
					waveobj.MetaKey_Url:  fileArg,
				},
			},
			Magnified: viewMagnified,
		}
	} else {
		absFile, err := filepath.Abs(fileArg)
		if err != nil {
			return fmt.Errorf("getting absolute path: %w", err)
		}
		absParent, err := filepath.Abs(filepath.Dir(fileArg))
		if err != nil {
			return fmt.Errorf("getting absolute path of parent dir: %w", err)
		}
		_, err = os.Stat(absParent)
		if err == fs.ErrNotExist {
			return fmt.Errorf("parent directory does not exist: %q", absParent)
		}
		if err != nil {
			return fmt.Errorf("getting file info: %w", err)
		}
		wshCmd = &wshrpc.CommandCreateBlockData{
			BlockDef: &waveobj.BlockDef{
				Meta: map[string]interface{}{
					waveobj.MetaKey_View: "preview",
					waveobj.MetaKey_File: absFile,
				},
			},
			Magnified: viewMagnified,
		}
		if cmd.Use == "edit" {
			wshCmd.BlockDef.Meta[waveobj.MetaKey_Edit] = true
		}
		if conn != "" {
			wshCmd.BlockDef.Meta[waveobj.MetaKey_Connection] = conn
		}
	}
	_, err := RpcClient.SendRpcRequest(wshrpc.Command_CreateBlock, wshCmd, &wshrpc.RpcOpts{Timeout: 2000})
	if err != nil {
		return fmt.Errorf("running view command: %w", err)
	}
	return nil
}
