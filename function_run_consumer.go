package bloc_client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fBloc/bloc-client-go/internal/event"
	"github.com/spf13/cast"
)

func (bC *blocClient) FunctionRunConsumer() {
	event.InjectMq(bC.GetOrCreateEventMQ())
	funcToRunEventChan := make(chan event.DomainEvent)
	err := event.ListenEvent(
		&event.ClientRunFunction{ClientName: bC.Name},
		bC.Name, funcToRunEventChan)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for functionToRunEvent := range funcToRunEventChan {
		wg.Add(1)
		go func(e event.DomainEvent, wg *sync.WaitGroup) {
			defer wg.Done()
			defer event.AckEvent(e)

			functionRunRecordIDStr := e.Identity()
			logger := bC.CreateFunctionRunLogger(functionRunRecordIDStr)

			funcRunRecordIns, err := bC.GetFunctionRunRecordByID(functionRunRecordIDStr)
			if err != nil {
				msg := fmt.Sprintf(
					"get function_run_record_ins by id-%s failed. error: %v",
					functionRunRecordIDStr, err)
				logger.Errorf(msg)
				funcRunOpt := NewFailedFunctionRunOpt(msg)
				bC.ReportFuncRunFinished(context.TODO(), functionRunRecordIDStr, *funcRunOpt)
				return
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
				return
			}

			// report function_run start
			err = bC.ReportFuncRunStart(traceCtx, functionRunRecordIDStr)
			if err != nil {
				logger.Errorf("report function run start to server failed: %v", err)
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
				return
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
					return
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

			// run the function
			go func() {
				functionIns.ExeFunc.Run(
					ctx, functionIns.Ipts,
					progressReportChan, functionRunOptChan,
					logger)
			}()

			// read the real-time msg & forward 2 server
			for {
				select {
				// 1. timeout
				case <-timeOutChan:
					logger.Infof("function run timeout canceled. function_run_record_id: %s", functionRunRecordIDStr)
					funcRunOpt = &FunctionRunOpt{
						Suc:             true,
						TimeoutCanceled: true}
					goto FunctionNodeRunFinished
				// 2. flow is canceled
				case <-cancelCheckTimer.C:
					isCanceled, err := bC.FlowRunIsCanceled(funcRunRecordIns.FlowRunRecordID)
					if err == nil && isCanceled {
						logger.Infof("function run is canceled from flow")
						funcRunOpt = &FunctionRunOpt{
							Suc:      true,
							Canceled: true}
						goto FunctionNodeRunFinished
					}
				// 3. report run progress
				case runningStatus := <-progressReportChan:
					bC.ReportFuncRunProgress(
						traceCtx,
						functionRunRecordIDStr, runningStatus.Progress,
						runningStatus.Msg, runningStatus.ProgressMilestoneIndex)
				// 4. finished!
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
				funcRunOpt.Brief = make(map[string]string, len(funcRunOpt.Detail))
				funcRunOpt.KeyMapObjectStorageKey = make(map[string]string, len(funcRunOpt.Detail))
				funcOptKeyMapValueType, funcOptKeyMapValueIsArray := functionIns.OptKeyMapValueTypeAndIsArray()
				for optKey, optVal := range funcRunOpt.Detail {
					minLength := 51
					valueType := funcOptKeyMapValueType[optKey]
					isArray := funcOptKeyMapValueIsArray[optKey]
					briefValue := ""
					if !isArray {
						switch valueType {
						case StringValueType, JsonValueType: // only truncate long string
							tmp, err := cast.ToStringE(optVal)
							if err == nil {
								tmpRune := []rune(tmp)
								rightIndex := minLength
								if len(tmpRune) < minLength {
									rightIndex = len(tmpRune)
								}
								briefValue = string(tmpRune[:rightIndex])
							}
						default:
							tmp, err := cast.ToStringE(optVal)
							if err == nil {
								briefValue = tmp
							}
						}
					}
					if briefValue != "" {
						funcRunOpt.Brief[optKey] = briefValue
					}

					serverPersisResp, err := bC.PersistFunctionRunOptFieldToServer(
						functionRunRecordIDStr, optKey, optVal)
					if err != nil {
						funcRunOpt.Brief[optKey] = "persist opt data to server failed: " + err.Error()
					} else {
						if _, ok := funcRunOpt.Brief[optKey]; !ok {
							funcRunOpt.Brief[optKey] = serverPersisResp.Brief
						}
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
		}(functionToRunEvent, &wg)

		wg.Wait()
	}
}
