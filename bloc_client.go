package bloc_client

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"sync"

	"github.com/fBloc/bloc-client-go/internal/conns/minio"
	"github.com/fBloc/bloc-client-go/internal/mq"
	"github.com/fBloc/bloc-client-go/internal/mq/rabbit"
	"github.com/fBloc/bloc-client-go/internal/object_storage"
	minioInf "github.com/fBloc/bloc-client-go/internal/object_storage/minio"
)

const serverBasicPathPrefix = "/api/v1/client/"

type BlocServerConfig struct {
	IP   string
	Port int
}

func (bSC *BlocServerConfig) IsNil() bool {
	if bSC == nil {
		return true
	}
	return bSC.IP == "" || bSC.Port == 0
}

func (bSC *BlocServerConfig) String() string {
	return fmt.Sprintf("%s:%d", bSC.IP, bSC.Port)
}

type RabbitConfig struct {
	User     string
	Password string
	Host     []string
	Vhost    string
}

func (rC *RabbitConfig) IsNil() bool {
	if rC == nil {
		return true
	}
	return len(rC.Host) <= 0 ||
		rC.User == "" ||
		rC.Password == ""
}

type MinioConfig struct {
	BucketName     string
	AccessKey      string
	AccessPassword string
	Addresses      []string
}

func (mF *MinioConfig) IsNil() bool {
	if mF == nil {
		return true
	}
	return mF.BucketName == "" || mF.AccessKey == "" ||
		mF.AccessPassword == "" || len(mF.Addresses) == 0
}

type ConfigBuilder struct {
	ServerConf *BlocServerConfig
	RabbitConf *RabbitConfig
	MinioConf  *MinioConfig
}

func (confbder *ConfigBuilder) SetServer(ip string, port int) *ConfigBuilder {
	confbder.ServerConf = &BlocServerConfig{IP: ip, Port: port}
	return confbder
}

func (confbder *ConfigBuilder) SetRabbitConfig(
	user, password string, host []string, vHost string,
) *ConfigBuilder {
	confbder.RabbitConf = &RabbitConfig{
		User:     user,
		Password: password,
		Host:     host,
		Vhost:    vHost}
	return confbder
}

func (confbder *ConfigBuilder) SetMinioConfig(
	bucketName string, addresses []string, key, password string) *ConfigBuilder {
	// minio名称不允许有下划线
	bucketName = strings.Replace(bucketName, "_", "", -1)
	confbder.MinioConf = &MinioConfig{
		BucketName:     bucketName,
		Addresses:      addresses,
		AccessKey:      key,
		AccessPassword: password}
	return confbder
}

func (congbder *ConfigBuilder) BuildUp() {
	// ServerConf http server 地址配置。
	if congbder.ServerConf.IsNil() {
		panic("must set bloc-server address")
	}

	// RabbitConf。需要检查输入的配置能够建立有效的链接
	if congbder.RabbitConf.IsNil() {
		panic("must set rabbit config")
	}
	rabbit.InitChannel((*rabbit.RabbitConfig)(congbder.RabbitConf))

	// MinioConf 如果输入了，需要查看minIO是否能够有效工作
	if !congbder.MinioConf.IsNil() {
		minio.Init((*minio.MinioConfig)(congbder.MinioConf))
	}
}

type Function struct {
	ID                 string
	Name               string
	GroupName          string
	Description        string
	Ipts               Ipts
	Opts               []*Opt
	ProgressMilestones []string
	ExeFunc            BlocFunctionNodeInterface
}

func (f *Function) IsNil() bool {
	return f == nil || f.ID == "" || f.ExeFunc == nil
}

type FunctionGroup struct {
	Name      string
	Functions []*Function
}

func (functionGroup *FunctionGroup) AddFunction(
	name string,
	description string,
	userImplementedFunc BlocFunctionNodeInterface) {
	for _, function := range functionGroup.Functions {
		if function.Name == name {
			errorInfo := fmt.Sprintf(
				"should not have same function name(%s) under same group(%s)",
				name, functionGroup.Name)
			panic(errorInfo)
		}
	}

	aggFunction := Function{
		Name:               name,
		GroupName:          functionGroup.Name,
		Description:        description,
		Ipts:               userImplementedFunc.IptConfig(),
		Opts:               userImplementedFunc.OptConfig(),
		ProgressMilestones: userImplementedFunc.AllProgressMilestones(),
		ExeFunc:            userImplementedFunc}

	functionGroup.Functions = append(functionGroup.Functions, &aggFunction)
}

type blocClient struct {
	Name           string
	FunctionGroups []*FunctionGroup
	configBuilder  *ConfigBuilder
	eventMQ        mq.MsgQueue
	objectStorage  object_storage.ObjectStorage
	sync.Mutex
}

func NewTestClient() *blocClient {
	return NewClient("local_test")
}

func NewClient(clientName string) *blocClient {
	return &blocClient{
		Name: clientName,
	}
}

func (bC *blocClient) CreateFunctionRunLogger(
	funcRunRecordID string,
) *Logger {
	return NewLogger(
		"func-run-record",
		bC.GenReqServerPath(), funcRunRecordID)
}

// GetConfigBuilder
func (bloc *blocClient) GetConfigBuilder() *ConfigBuilder {
	bloc.configBuilder = &ConfigBuilder{}
	return bloc.configBuilder
}

func (bloc *blocClient) RegisterFunctionGroup(
	name string,
) *FunctionGroup {
	for _, i := range bloc.FunctionGroups {
		if i.Name == name {
			panic("should not register same name group")
		}
	}
	functionGroup := FunctionGroup{
		Name: name, Functions: make([]*Function, 0, 10),
	}
	bloc.FunctionGroups = append(bloc.FunctionGroups, &functionGroup)
	return &functionGroup
}

func (bC *blocClient) GetOrCreateEventMQ() mq.MsgQueue {
	bC.Lock()
	defer bC.Unlock()
	if bC.eventMQ != nil {
		return bC.eventMQ
	}

	rabbitMQ := rabbit.InitChannel(
		(*rabbit.RabbitConfig)(bC.configBuilder.RabbitConf))
	bC.eventMQ = rabbitMQ

	return bC.eventMQ
}

func (bC *blocClient) GetFunctionByID(functionID string) Function {
	for _, fGroup := range bC.FunctionGroups {
		for _, f := range fGroup.Functions {
			if f.ID == functionID {
				return *f
			}
		}
	}
	return Function{}
}

func (bC *blocClient) GetOrCreateObjectStorage() object_storage.ObjectStorage {
	if bC.objectStorage != nil {
		return bC.objectStorage
	}
	bC.Lock()
	defer bC.Unlock()

	minioOS := minioInf.New(
		bC.configBuilder.MinioConf.Addresses,
		bC.configBuilder.MinioConf.AccessKey,
		bC.configBuilder.MinioConf.AccessPassword,
		bC.configBuilder.MinioConf.BucketName,
	)
	bC.objectStorage = minioOS

	return bC.objectStorage
}

func (bC *blocClient) GenReqServerPath(subPaths ...string) string {
	resp := path.Join(
		bC.configBuilder.ServerConf.String(),
		serverBasicPathPrefix)
	for _, subPath := range subPaths {
		resp = path.Join(resp, subPath)
	}
	return resp
}

func (bC *blocClient) TestRunFunction(
	userFunction BlocFunctionNodeInterface,
	iptValues [][]interface{},
) FunctionRunOpt {
	progressReportChan := make(chan HighReadableFunctionRunProgress)
	functionRunOptChan := make(chan *FunctionRunOpt)
	logger := newMockLogger()

	userFunctionIpts := userFunction.IptConfig()
	for iptIndex, i := range userFunctionIpts {
		must := userFunctionIpts[iptIndex].Must
		if must {
			if len(iptValues)-1 < iptIndex {
				panic(fmt.Sprintf(
					"index %d is a cannot be nil ipt, but params has no this data",
					iptIndex))
			}
			if len(iptValues[iptIndex]) < len(userFunctionIpts[iptIndex].Components) {
				panic(
					fmt.Sprintf("index %d need %d component value, ", iptIndex, len(userFunctionIpts[iptIndex].Components)) +
						fmt.Sprintf("but param only provide %d value", len(iptValues[iptIndex])))
			}
		}

		if iptIndex >= len(iptValues) {
			break
		}
		for componentIndex := range i.Components {
			userFunctionIpts[iptIndex].Components[componentIndex].Value = iptValues[iptIndex][componentIndex]
		}
	}

	go func() {
		userFunction.Run(
			context.TODO(),
			userFunctionIpts,
			progressReportChan,
			functionRunOptChan,
			logger)
	}()

	for {
		select {
		// function运行进度上报
		case runningStatus := <-progressReportChan:
			log.Printf("reporting progress: %v", runningStatus)
		// 运行成功完成
		case funcRunOpt := <-functionRunOptChan:
			log.Printf("run finished with resp: %+v", funcRunOpt)
			return *funcRunOpt
		}
	}
}
