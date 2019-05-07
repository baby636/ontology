/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package server

import (
	"fmt"
	common2 "github.com/ontio/ontology/common"
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver"
	"github.com/ontio/ontology/p2pserver/common"
)

type P2PActor struct {
	props  *actor.Props
	server *p2pserver.P2PServer
}

// NewP2PActor creates an actor to handle the messages from
// txnpool and consensus
func NewP2PActor(p2pServer *p2pserver.P2PServer) *P2PActor {
	return &P2PActor{
		server: p2pServer,
	}
}

//start a actor called net_server
func (this *P2PActor) Start(shardID common2.ShardID) (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	p2pPid, err := actor.SpawnNamed(this.props, "net_server"+fmt.Sprintf("%d", shardID.ToUint64()))
	return p2pPid, err
}

//message handler
func (this *P2PActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[p2p]actor restarting")
	case *actor.Stopping:
		log.Warn("[p2p]actor stopping")
	case *actor.Stopped:
		log.Warn("[p2p]actor stopped")
	case *actor.Started:
		log.Debug("[p2p]actor started")
	case *actor.Restart:
		log.Warn("[p2p]actor restart")
	case *StopServerReq:
		this.handleStopServerReq(ctx, msg)
	case *GetPortReq:
		this.handleGetPortReq(ctx, msg)
	case *GetVersionReq:
		this.handleGetVersionReq(ctx, msg)
	case *GetConnectionCntReq:
		this.handleGetConnectionCntReq(ctx, msg)
	case *GetIdReq:
		this.handleGetIDReq(ctx, msg)
	case *GetConnectionStateReq:
		this.handleGetConnectionStateReq(ctx, msg)
	case *GetTimeReq:
		this.handleGetTimeReq(ctx, msg)
	case *GetNeighborAddrsReq:
		this.handleGetNeighborAddrsReq(ctx, msg)
	case *GetRelayStateReq:
		this.handleGetRelayStateReq(ctx, msg)
	case *GetNodeTypeReq:
		this.handleGetNodeTypeReq(ctx, msg)
	case *TransmitConsensusMsgReq:
		this.handleTransmitConsensusMsgReq(ctx, msg)
	case *common.AppendPeerID:
		this.server.OnAddNode(msg.ID, msg.Shards)
	case *common.RemovePeerID:
		this.server.OnDelNode(msg.ID, msg.Shards)
	case *common.AppendHeaders:
		this.server.OnHeaderReceive(msg.FromID, msg.Headers)
	case *common.AppendBlock:
		this.server.OnBlockReceive(msg.FromID, msg.BlockSize, msg.Block, msg.MerkleRoot)
	case *StartSync:
		if shardID, err := common2.NewShardID(msg.ShardID); err == nil {
			if err := this.server.StartSyncShard(shardID); err != nil {
				log.Errorf("[p2p] start syncing shard %d: %s", shardID, err)
			}
		} else {
			log.Errorf("[p2p] start syncing, invalid shardID: %d", msg.ShardID)
		}
	default:
		err := this.server.Xmit(ctx.Message())
		if nil != err {
			log.Warn("[p2p]error xmit message ", err.Error(), reflect.TypeOf(ctx.Message()))
		}
	}
}

//stop handler
func (this *P2PActor) handleStopServerReq(ctx actor.Context, req *StopServerReq) {
	this.server.Stop()
	if ctx.Sender() != nil {
		resp := &StopServerRsp{}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//get port handler
func (this *P2PActor) handleGetPortReq(ctx actor.Context, req *GetPortReq) {
	syncPort := this.server.GetPort()
	if ctx.Sender() != nil {
		resp := &GetPortRsp{
			SyncPort: syncPort,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//version handler
func (this *P2PActor) handleGetVersionReq(ctx actor.Context, req *GetVersionReq) {
	version := this.server.GetVersion()
	if ctx.Sender() != nil {
		resp := &GetVersionRsp{
			Version: version,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//connection count handler
func (this *P2PActor) handleGetConnectionCntReq(ctx actor.Context, req *GetConnectionCntReq) {
	cnt := this.server.GetConnectionCnt()
	if ctx.Sender() != nil {
		resp := &GetConnectionCntRsp{
			Cnt: cnt,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//get id handler
func (this *P2PActor) handleGetIDReq(ctx actor.Context, req *GetIdReq) {
	id := this.server.GetID()
	if ctx.Sender() != nil {
		resp := &GetIdRsp{
			Id: id,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//connection state handler
func (this *P2PActor) handleGetConnectionStateReq(ctx actor.Context, req *GetConnectionStateReq) {
	state := this.server.GetConnectionState()
	if ctx.Sender() != nil {
		resp := &GetConnectionStateRsp{
			State: state,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//timestamp handler
func (this *P2PActor) handleGetTimeReq(ctx actor.Context, req *GetTimeReq) {
	time := this.server.GetTime()
	if ctx.Sender() != nil {
		resp := &GetTimeRsp{
			Time: time,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//nbr peer`s address handler
func (this *P2PActor) handleGetNeighborAddrsReq(ctx actor.Context, req *GetNeighborAddrsReq) {
	addrs := this.server.GetNeighborAddrs()
	if ctx.Sender() != nil {
		resp := &GetNeighborAddrsRsp{
			Addrs: addrs,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//peer`s relay state handler
func (this *P2PActor) handleGetRelayStateReq(ctx actor.Context, req *GetRelayStateReq) {
	ret := this.server.GetNetWork().GetRelay()
	if ctx.Sender() != nil {
		resp := &GetRelayStateRsp{
			Relay: ret,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//peer`s service type handler
func (this *P2PActor) handleGetNodeTypeReq(ctx actor.Context, req *GetNodeTypeReq) {
	ret := this.server.GetNetWork().GetServices()
	if ctx.Sender() != nil {
		resp := &GetNodeTypeRsp{
			NodeType: ret,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (this *P2PActor) handleTransmitConsensusMsgReq(ctx actor.Context, req *TransmitConsensusMsgReq) {
	peer := this.server.GetNetWork().GetPeer(req.Target)
	if peer != nil {
		this.server.Send(peer, req.Msg, true)
	} else {
		log.Warnf("[p2p]can`t transmit consensus msg:no valid neighbor peer: %d\n", req.Target)
	}
}
