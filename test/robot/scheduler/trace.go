// Copyright (C) 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduler

import (
	"context"

	"github.com/google/gapid/core/log"
	"github.com/google/gapid/test/robot/build"
	"github.com/google/gapid/test/robot/job"
	"github.com/google/gapid/test/robot/monitor"
	"github.com/google/gapid/test/robot/trace"
)

func (s schedule) getTraceTargetTools(ctx context.Context, subj *monitor.Subject) *build.AndroidToolSet {
	ctx = log.V{"target": s.worker.Target}.Bind(ctx)
	tools := s.pkg.FindToolsForAPK(ctx, s.data.FindDevice(s.worker.Host), s.data.FindDevice(s.worker.Target), subj.GetAPK())

	if tools == nil {
		return nil
	}
	if tools.GapidApk == "" {
		return nil
	}
	return tools
}

func (s schedule) doTrace(ctx context.Context, subj *monitor.Subject) error {
	if !s.worker.Supports(job.Trace) {
		return nil
	}
	ctx = log.Enter(ctx, "Trace")
	ctx = log.V{"Package": s.pkg.Id}.Bind(ctx)
	hostTools := s.getHostTools(ctx)
	targetTools := s.getTraceTargetTools(ctx, subj)
	if hostTools == nil || targetTools == nil {
		return log.Err(ctx, nil, "Failed to find tools for trace!")
	}
	input := &trace.Input{
		Subject:  subj.Id,
		Gapit:    hostTools.Host.Gapit,
		GapidApk: targetTools.GapidApk,
		Hints:    subj.Hints,
		Layout: &trace.ToolingLayout{
			GapidAbi: targetTools.Abi,
		},
	}
	action := &trace.Action{
		Input:  input,
		Host:   s.worker.Host,
		Target: s.worker.Target,
	}
	if _, found := s.data.Traces.FindOrCreate(ctx, action); found {
		return nil
	}
	// TODO: we just ignore the error right now, what should we do?
	go s.managers.Trace.Do(ctx, action.Target, input)
	return nil
}
