// Copyright 2017 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package sql

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/storage/engine/enginepb"
	"github.com/pkg/errors"
)

// SetTransaction sets a transaction's isolation level
func (p *planner) SetTransaction(n *tree.SetTransaction) (planNode, error) {
	return &zeroNode{}, p.setTransactionModes(n.Modes)
}

func (p *planner) setTransactionModes(modes tree.TransactionModes) error {
	if err := p.setIsolationLevel(modes.Isolation); err != nil {
		return err
	}
	if err := p.setUserPriority(modes.UserPriority); err != nil {
		return err
	}
	return p.setReadWriteMode(modes.ReadWriteMode)
}

func (p *planner) setIsolationLevel(level tree.IsolationLevel) error {
	var iso enginepb.IsolationType
	switch level {
	case tree.UnspecifiedIsolation:
		return nil
	case tree.SnapshotIsolation:
		iso = enginepb.SNAPSHOT
	case tree.SerializableIsolation:
		iso = enginepb.SERIALIZABLE
	default:
		return errors.Errorf("unknown isolation level: %s", level)
	}

	return p.session.TxnState.setIsolationLevel(iso)
}

func (p *planner) setUserPriority(userPriority tree.UserPriority) error {
	var up roachpb.UserPriority
	switch userPriority {
	case tree.UnspecifiedUserPriority:
		return nil
	case tree.Low:
		up = roachpb.MinUserPriority
	case tree.Normal:
		up = roachpb.NormalUserPriority
	case tree.High:
		up = roachpb.MaxUserPriority
	default:
		return errors.Errorf("unknown user priority: %s", userPriority)
	}
	return p.session.TxnState.setPriority(up)
}

// Note: This setting currently doesn't have any effect and therefor is not
// persisted anywhere. If this changes, care needs to be taken to restore it
// when ROLLBACK TO SAVEPOINT starts a new sql transaction.
func (p *planner) setReadWriteMode(readWriteMode tree.ReadWriteMode) error {
	switch readWriteMode {
	case tree.UnspecifiedReadWriteMode:
		return nil
	case tree.ReadOnly:
		return errors.New("read only not supported")
	case tree.ReadWrite:
		return nil
	default:
		return errors.Errorf("unknown read mode: %s", readWriteMode)
	}
}
