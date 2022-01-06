/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scheduler

import (
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/base/common"
)

func NewZeroPieceResult(taskID, peerID string) *PieceResult {
	return &PieceResult{
		TaskId: taskID,
		SrcPid: peerID,
		PieceInfo: &base.PieceInfo{
			PieceNum:    common.ZeroOfPiece,
			RangeStart:  0,
			RangeSize:   0,
			PieceMd5:    "",
			PieceOffset: 0,
			PieceStyle:  0,
		},
	}
}

func NewEndPieceResult(taskID, peerID string, finishedCount int32) *PieceResult {
	return &PieceResult{
		TaskId:        taskID,
		SrcPid:        peerID,
		FinishedCount: finishedCount,
		PieceInfo: &base.PieceInfo{
			PieceNum:    common.EndOfPiece,
			RangeStart:  0,
			RangeSize:   0,
			PieceMd5:    "",
			PieceOffset: 0,
			PieceStyle:  0,
		},
	}
}

func NewQueryPieceResult(taskID, peerID string) *PieceResult {
	return &PieceResult{
		TaskId:        taskID,
		SrcPid:        peerID,
		FinishedCount: 0,
		PieceInfo: &base.PieceInfo{
			PieceNum:    common.QueryOfPiece,
			RangeStart:  0,
			RangeSize:   0,
			PieceMd5:    "",
			PieceOffset: 0,
			PieceStyle:  0,
		},
	}
}
