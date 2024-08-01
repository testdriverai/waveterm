// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package clientservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wavetermdev/thenextwave/pkg/eventbus"
	"github.com/wavetermdev/thenextwave/pkg/service/objectservice"
	"github.com/wavetermdev/thenextwave/pkg/util/utilfn"
	"github.com/wavetermdev/thenextwave/pkg/wstore"
)

type ClientService struct{}

const DefaultTimeout = 2 * time.Second

func (cs *ClientService) GetClientData() (*wstore.Client, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	clientData, err := wstore.DBGetSingleton[*wstore.Client](ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client data: %w", err)
	}
	return clientData, nil
}

func (cs *ClientService) GetWorkspace(workspaceId string) (*wstore.Workspace, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ws, err := wstore.DBGet[*wstore.Workspace](ctx, workspaceId)
	if err != nil {
		return nil, fmt.Errorf("error getting workspace: %w", err)
	}
	return ws, nil
}

func (cs *ClientService) GetTab(tabId string) (*wstore.Tab, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	tab, err := wstore.DBGet[*wstore.Tab](ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error getting tab: %w", err)
	}
	return tab, nil
}

func (cs *ClientService) GetWindow(windowId string) (*wstore.Window, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	window, err := wstore.DBGet[*wstore.Window](ctx, windowId)
	if err != nil {
		return nil, fmt.Errorf("error getting window: %w", err)
	}
	return window, nil
}

func (cs *ClientService) MakeWindow(ctx context.Context) (*wstore.Window, error) {
	return wstore.CreateWindow(ctx)
}

// moves the window to the front of the windowId stack
func (cs *ClientService) FocusWindow(ctx context.Context, windowId string) error {
	client, err := cs.GetClientData()
	if err != nil {
		return err
	}
	winIdx := utilfn.SliceIdx(client.WindowIds, windowId)
	if winIdx == -1 {
		return nil
	}
	client.WindowIds = utilfn.MoveSliceIdxToFront(client.WindowIds, winIdx)
	return wstore.DBUpdate(ctx, client)
}

func (cs *ClientService) AgreeTos(ctx context.Context) (wstore.UpdatesRtnType, error) {
	ctx = wstore.ContextWithUpdates(ctx)
	clientData, err := wstore.DBGetSingleton[*wstore.Client](ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client data: %w", err)
	}
	timestamp := time.Now().UnixMilli()
	clientData.TosAgreed = timestamp
	err = wstore.DBUpdate(ctx, clientData)
	if err != nil {
		return nil, fmt.Errorf("error updating client data: %w", err)
	}
	cs.BootstrapStarterLayout(ctx)
	return wstore.ContextGetUpdatesRtn(ctx), nil
}

type PortableLayout []struct {
	IndexArr []int
	Size     uint
	BlockDef *wstore.BlockDef
}

func (cs *ClientService) BootstrapStarterLayout(ctx context.Context) error {
	ctx, cancelFn := context.WithTimeout(ctx, 2*time.Second)
	defer cancelFn()
	client, err := wstore.DBGetSingleton[*wstore.Client](ctx)
	if err != nil {
		log.Printf("unable to find client: %v\n", err)
		return fmt.Errorf("unable to find client: %w", err)
	}

	if len(client.WindowIds) < 1 {
		return fmt.Errorf("error bootstrapping layout, no windows exist")
	}

	windowId := client.WindowIds[0]

	window, err := wstore.DBMustGet[*wstore.Window](ctx, windowId)
	if err != nil {
		return fmt.Errorf("error getting window: %w", err)
	}

	tabId := window.ActiveTabId

	starterLayout := PortableLayout{
		{IndexArr: []int{0}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View:       "term",
				wstore.MetaKey_Controller: "shell",
			},
		}},
		{IndexArr: []int{1}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View: "cpuplot",
			},
		}},
		{IndexArr: []int{1, 1}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View: "web",
				wstore.MetaKey_Url:  "https://github.com/wavetermdev/waveterm",
			},
		}},
		{IndexArr: []int{1, 2}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View: "preview",
				wstore.MetaKey_File: "~",
			},
		}},
		{IndexArr: []int{2}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View:       "term",
				wstore.MetaKey_Controller: "shell",
			},
		}},
		{IndexArr: []int{2, 1}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View: "waveai",
			},
		}},
		{IndexArr: []int{2, 2}, BlockDef: &wstore.BlockDef{
			Meta: wstore.MetaMapType{
				wstore.MetaKey_View: "web",
				wstore.MetaKey_Url:  "https://www.youtube.com/embed/cKqsw_sAsU8",
			},
		}},
	}

	objsvc := &objectservice.ObjectService{}

	for i := 0; i < len(starterLayout); i++ {
		layoutAction := starterLayout[i]

		blockData, err := objsvc.CreateBlock_NoUI(ctx, tabId, layoutAction.BlockDef, &wstore.RuntimeOpts{})

		if err != nil {
			return fmt.Errorf("unable to create block for starter layout: %w", err)
		}

		eventbus.SendEventToWindow(windowId, eventbus.WSEventType{
			EventType: eventbus.WSEvent_LayoutAction,
			Data: &eventbus.WSLayoutActionData{
				ActionType: "insertatindex",
				TabId:      tabId,
				BlockId:    blockData.OID,
				IndexArr:   layoutAction.IndexArr,
				NodeSize:   layoutAction.Size,
			},
		})
	}
	return nil
}
