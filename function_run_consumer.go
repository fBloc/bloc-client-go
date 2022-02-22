package bloc_client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fBloc/bloc-client-go/internal/event"
)

func (bC *BlocClient) FunctionRunConsumer() {
	event.InjectMq(bC.GetOrCreateEventMQ())
	objectStorage := bC.GetOrCreateObjectStorage()
	funcToRunEventChan := make(chan event.DomainEvent)
	err := event.ListenEvent(
		&event.ClientRunFunction{ClientName: bC.Name},
		bC.Name, funcToRunEventChan)
	if err != nil {
		panic(err)
	}

	for functionToRunEvent := range funcToRunEventChan {
		functionRunRecordIDStr := functionToRunEvent.Identity()
		funcRunRecordIns, err := bC.GetFunctionRunRecordByID(functionRunRecordIDStr)
		if err != nil {
			// TODO
		}
		logger := bC.CreateFunctionRunLogger(functionRunRecordIDStr)

		// make sure you copied functionIns! donnot disrupt the oringin functionIns
		functionIns := bC.GetFunctionByID(funcRunRecordIns.FunctionID)
		if functionIns.IsNil() {
			// TODO
		}

		// 从brief中恢复出完整的ipt以供运行
		for iptIndex, ipt := range funcRunRecordIns.IptBriefAndObjectStoragekey {
			for componentIndex, componentBrief := range ipt {
				dataByte, err := objectStorage.Get(componentBrief.ObjectStorageKey)
				if err != nil {
					// TODO
				}
				var data interface{}
				err = json.Unmarshal(dataByte, &data)
				if err != nil {
					// TODO
				}

				functionIns.Ipts[iptIndex].Components[componentIndex].Value = data
			}
		}

		// 超时检测
		timeOutChan := make(chan struct{})
		if !funcRunRecordIns.ShouldBeCanceledAt.IsZero() { // 设置了整体运行的超时时长
			if funcRunRecordIns.ShouldBeCanceledAt.Before(time.Now()) { // 已超时
				// TODO 触发上报已超时、不运行
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

		// 开始运行
		go func() {
			functionIns.ExeFunc.Run(
				ctx, functionIns.Ipts,
				progressReportChan, functionRunOptChan,
				logger)
		}()
		for {
			select {
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
					functionRunRecordIDStr,
					runningStatus.Progress,
					runningStatus.Msg,
					runningStatus.ProcessStageIndex)
			// 4. 运行成功完成
			case funcRunOpt = <-functionRunOptChan:
				logger.Infof("function run suc")
				goto FunctionNodeRunFinished
			}
		}
	FunctionNodeRunFinished:
		cancelFunctionExecute()
		close(progressReportChan)
		cancelCheckTimer.Stop()
		// save opt
		if funcRunOpt.Suc {
			funcRunOpt.KeyMapObjectStorageKey = make(map[string]string, len(funcRunOpt.Detail))
			for optKey, optVal := range funcRunOpt.Detail {
				uploadByte, _ := json.Marshal(optVal)
				ossKey := functionRunRecordIDStr + "_" + optKey
				err = objectStorage.Set(ossKey, uploadByte)
				if err == nil {
					optInRune := []rune(string(uploadByte))
					minLength := 51
					if len(optInRune) < minLength {
						minLength = len(optInRune)
					}
					if funcRunOpt.Brief == nil {
						funcRunOpt.Brief = make(map[string]string, len(funcRunOpt.Detail))
					}
					funcRunOpt.Brief[optKey] = string(optInRune[:minLength-1])
					funcRunOpt.KeyMapObjectStorageKey[optKey] = ossKey
				} else {
					funcRunOpt.Brief[optKey] = "存储运行输出到对象存储失败"
				}
			}
		}

		// report finished
		err = bC.ReportFuncRunFinished(functionRunRecordIDStr, *funcRunOpt)
		if err != nil {
			logger.Errorf("report function run finished failed: %+v", err)
		}
	}
}
