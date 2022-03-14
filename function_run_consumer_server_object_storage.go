package bloc_client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fBloc/bloc-client-go/internal/event"
)

func (bC *blocClient) FunctionRunConsumerWithoutLocalObjectStorageImplemention() {
	event.InjectMq(bC.GetOrCreateEventMQ())
	funcToRunEventChan := make(chan event.DomainEvent)
	err := event.ListenEvent(
		&event.ClientRunFunction{ClientName: bC.Name},
		bC.Name, funcToRunEventChan)
	if err != nil {
		panic(err)
	}

	for functionToRunEvent := range funcToRunEventChan {
		functionRunRecordIDStr := functionToRunEvent.Identity()
		logger := bC.CreateFunctionRunLogger(functionRunRecordIDStr)

		funcRunRecordIns, err := bC.GetFunctionRunRecordByID(functionRunRecordIDStr)
		if err != nil {
			msg := fmt.Sprintf(
				"get function_run_record_ins by id-%s failed. error: %v",
				functionRunRecordIDStr, err)
			logger.Errorf(msg)
			funcRunOpt := NewFailedFunctionRunOpt(msg)
			bC.ReportFuncRunFinished(context.TODO(), functionRunRecordIDStr, *funcRunOpt)
			continue
		}

		spanID := NewSpanID()
		logger.SetTraceIDAndSpanID(funcRunRecordIns.TraceID, spanID)
		logger.Infof("set trace_id: %s, spanID: %s", funcRunRecordIns.TraceID, spanID)

		traceCtx := SetTraceIDAndSpanIDToContext(funcRunRecordIns.TraceID, spanID)
		// make sure you copied functionIns! donnot disrupt the oringin functionIns
		functionIns := bC.GetFunctionByID(funcRunRecordIns.FunctionID)
		if functionIns.IsNil() {
			msg := fmt.Sprintf(
				"get function_ins by id-%s failed", funcRunRecordIns.FunctionID)
			logger.Errorf(msg)
			funcRunOpt := NewFailedFunctionRunOpt(msg)
			bC.ReportFuncRunFinished(traceCtx, functionRunRecordIDStr, *funcRunOpt)
			continue
		}

		// 从brief中恢复出完整的ipt以供运行
		completeIptSuc := true
		for iptIndex, ipt := range funcRunRecordIns.IptBriefAndObjectStoragekey {
			for componentIndex, componentBrief := range ipt {
				dataByte, err := bC.FetchObjectStorageDataByKeyFromServer(componentBrief.ObjectStorageKey)
				if err != nil {
					msg := fmt.Sprintf(
						"get ipt value from objectStorage failed. iptIndex-%d, componentIndex-%d. componentBrief-%s. error: %v",
						iptIndex, componentIndex, componentBrief, err)
					logger.Errorf(msg)
					funcRunOpt := NewFailedFunctionRunOpt(msg)
					bC.ReportFuncRunFinished(traceCtx, functionRunRecordIDStr, *funcRunOpt)
					completeIptSuc = false
					break
				}

				var data interface{}
				err = json.Unmarshal(dataByte, &data)
				if err != nil {
					msg := fmt.Sprintf(
						"get ipt value from objectStorage suc, but json unmarshal it failed. iptIndex-%d, componentIndex-%d. componentBrief-%s. resp-string: %s. error: %v",
						iptIndex, componentIndex, componentBrief, string(dataByte), err)
					logger.Errorf(msg)
					funcRunOpt := NewFailedFunctionRunOpt(msg)
					bC.ReportFuncRunFinished(traceCtx, functionRunRecordIDStr, *funcRunOpt)
					completeIptSuc = false
					break
				}

				functionIns.Ipts[iptIndex].Components[componentIndex].Value = data
			}
		}
		if !completeIptSuc {
			continue
		}

		// 超时检测
		timeOutChan := make(chan struct{})
		if !funcRunRecordIns.ShouldBeCanceledAt.IsZero() { // 设置了整体运行的超时时长
			if funcRunRecordIns.ShouldBeCanceledAt.Before(time.Now()) { // 已超时
				msg := fmt.Sprintf(
					"already timeout. timeout time is: %s, now is: %s",
					funcRunRecordIns.ShouldBeCanceledAt.Format(time.RFC3339),
					time.Now().Format(time.RFC3339))
				logger.Errorf(msg)
				funcRunOpt := NewTimeoutCanceldFunctionRunOpt()
				bC.ReportFuncRunFinished(traceCtx, functionRunRecordIDStr, *funcRunOpt)
				continue
			} else { // 未超时
				timer := time.After(time.Until(funcRunRecordIns.ShouldBeCanceledAt))
				go func() {
					for range timer {
						timeOutChan <- struct{}{}
					}
				}()
			}
		}

		cancelCheckTimer := time.NewTicker(6 * time.Second)
		progressReportChan := make(chan HighReadableFunctionRunProgress)
		functionRunOptChan := make(chan *FunctionRunOpt)
		var funcRunOpt *FunctionRunOpt
		ctx := context.Background()
		ctx, cancelFunctionExecute := context.WithCancel(ctx)
		heartBeatTicker := time.NewTicker(5 * time.Second)

		// 开始运行
		go func() {
			functionIns.ExeFunc.Run(
				ctx, functionIns.Ipts,
				progressReportChan, functionRunOptChan,
				logger)
		}()
		for {
			select {
			// 0. 存活心跳上报
			case <-heartBeatTicker.C:
				bC.ReportFuncExecuteHeartbeat(
					traceCtx,
					functionRunRecordIDStr)
			// 1. 超时
			case <-timeOutChan:
				logger.Infof("function run timeout canceled. function_run_record_id: %s", functionRunRecordIDStr)
				funcRunOpt = &FunctionRunOpt{
					Suc:             true,
					TimeoutCanceled: true}
				goto FunctionNodeRunFinished
			// 2. flow被用户在前端取消
			case <-cancelCheckTimer.C:
				isCanceled, err := bC.FlowRunIsCanceled(funcRunRecordIns.FlowRunRecordID)
				if err == nil && isCanceled {
					logger.Infof("function run is canceled from flow")
					funcRunOpt = &FunctionRunOpt{
						Suc:      true,
						Canceled: true}
					goto FunctionNodeRunFinished
				}
			// 3. function运行进度上报
			case runningStatus := <-progressReportChan:
				bC.ReportFuncRunProgress(
					traceCtx,
					functionRunRecordIDStr, runningStatus.Progress,
					runningStatus.Msg, runningStatus.ProcessStageIndex)
			// 4. 运行成功完成
			case funcRunOpt = <-functionRunOptChan:
				logger.Infof("function run suc")
				goto FunctionNodeRunFinished
			}
		}
	FunctionNodeRunFinished:
		heartBeatTicker.Stop()
		cancelFunctionExecute()
		close(progressReportChan)
		cancelCheckTimer.Stop()

		// save opt
		if funcRunOpt.Suc {
			funcRunOpt.Brief = make(map[string]string, len(funcRunOpt.Detail))
			funcRunOpt.KeyMapObjectStorageKey = make(map[string]string, len(funcRunOpt.Detail))
			for optKey, optVal := range funcRunOpt.Detail {
				serverPersisResp, err := bC.PersistFunctionRunOptFieldToServer(
					functionRunRecordIDStr, optKey, optVal)
				if err != nil {
					funcRunOpt.Brief[optKey] = "persist opt data to server failed: " + err.Error()
				} else {
					funcRunOpt.Brief[optKey] = serverPersisResp.Brief
					funcRunOpt.KeyMapObjectStorageKey[optKey] = serverPersisResp.ObjectStorageKey
				}
			}
		}

		// report finished
		err = bC.ReportFuncRunFinished(traceCtx, functionRunRecordIDStr, *funcRunOpt)
		if err != nil {
			logger.Errorf("report function run finished failed: %+v", err)
		} else {
			logger.Infof("report function run finished suc")
		}
	}
}
