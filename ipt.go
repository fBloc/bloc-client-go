package bloc_client

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type IptComponent struct {
	ValueType       ValueType       `json:"value_type"`
	FormControlType FormControlType `json:"formcontrol_type"`
	Hint            string          `json:"hint"`
	DefaultValue    interface{}     `json:"default_value"`
	AllowMulti      bool            `json:"allow_multi"`
	SelectOptions   []SelectOption  `json:"select_options"` // only exist when FormControlType is selection
	Value           interface{}     `json:"-"`
}

func (ipt *IptComponent) String() string {
	resp := string(ipt.ValueType) + string(ipt.FormControlType) + strconv.FormatBool(ipt.AllowMulti)
	for _, option := range ipt.SelectOptions {
		resp = resp + fmt.Sprintf("%v", option.Value)
	}
	return resp
}

func (ipt *IptComponent) Config() map[string]interface{} {
	config := make(map[string]interface{}, 6)
	config["value_type"] = ipt.ValueType
	config["formControl_type"] = ipt.FormControlType
	config["hint"] = ipt.Hint
	config["value"] = ipt.DefaultValue
	config["allow_multi"] = ipt.AllowMulti
	config["options"] = ipt.SelectOptions
	return config
}

type Ipt struct {
	Key        string          `json:"key"`
	Display    string          `json:"display"`
	Must       bool            `json:"must"`
	Components []*IptComponent `json:"components"`
}

func (ipt *Ipt) String() string {
	resp := ipt.Key + ipt.Display + strconv.FormatBool(ipt.Must)
	for _, component := range ipt.Components {
		resp += component.String()
	}
	return resp
}

func (ipt *Ipt) Config() map[string]interface{} {
	config := make(map[string]interface{}, 4)
	config["key"] = ipt.Key
	config["display"] = ipt.Display
	config["must"] = ipt.Must

	fieldsConfigs := make([]map[string]interface{}, 0, len(ipt.Components))
	for _, field := range ipt.Components {
		fieldsConfigs = append(fieldsConfigs, field.Config())
	}
	config["components"] = fieldsConfigs
	return config
}

func (ipt *Ipt) GetIntValue(componentIndex int) (int, error) {
	if len(ipt.Components) < componentIndex-1 {
		return 0, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != IntValueType && component.ValueType != FloatValueType {
		return 0, errors.Errorf("valueType should be int/float but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return 0, errors.New("value should be a intSlice")
	}
	return cast.ToInt(component.Value), nil
}

func (ipt *Ipt) GetIntSliceValue(componentIndex int) (resp []int, err error) {
	if len(ipt.Components) < componentIndex-1 {
		return []int{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != IntValueType {
		return []int{}, errors.Errorf("valueType should be int but get: %s", component.ValueType)
	}
	resp, err = cast.ToIntSliceE(component.Value)
	return
}

func (ipt *Ipt) GetFloat64Value(componentIndex int) (float64, error) {
	if len(ipt.Components) < componentIndex-1 {
		return 0, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != FloatValueType {
		return 0, errors.Errorf("valueType should be float but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return 0, errors.New("value should be a intSlice")
	}
	return cast.ToFloat64(component.Value), nil
}

func (ipt *Ipt) GetFloat64SliceValue(componentIndex int) ([]float64, error) {
	if len(ipt.Components) < componentIndex-1 {
		return []float64{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != FloatValueType {
		return []float64{}, errors.Errorf("valueType should be float but get: %s", component.ValueType)
	}
	interfaceSlice := cast.ToSlice(component.Value)
	resp := make([]float64, 0, len(interfaceSlice))
	for _, i := range interfaceSlice {
		resp = append(resp, cast.ToFloat64(i))
	}
	return resp, nil
}

func (ipt *Ipt) GetStringValue(componentIndex int) (string, error) {
	if len(ipt.Components) < componentIndex-1 {
		return "", errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != StringValueType {
		return "", errors.Errorf("valueType should be string but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return "", errors.New("value should be a stringSlice")
	}
	return cast.ToString(component.Value), nil
}

func (ipt *Ipt) GetStringSliceValue(componentIndex int) (resp []string, err error) {
	if len(ipt.Components) < componentIndex-1 {
		return []string{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != StringValueType {
		return []string{}, errors.Errorf("valueType should be string but get: %s", component.ValueType)
	}
	if !component.AllowMulti {
		return []string{cast.ToString(component.Value)}, nil
	}
	resp, err = cast.ToStringSliceE(component.Value)
	return
}

func (ipt *Ipt) GetBoolValue(componentIndex int) (bool, error) {
	if len(ipt.Components) < componentIndex-1 {
		return false, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != BoolValueType {
		return false, errors.Errorf("valueType should be bool but get: %s", component.ValueType)
	}
	if component.AllowMulti {
		return false, errors.New("value should be a stringSlice")
	}
	return cast.ToBool(component.Value), nil
}

func (ipt *Ipt) GetBoolSliceValue(componentIndex int) (resp []bool, err error) {
	if len(ipt.Components) < componentIndex-1 {
		return []bool{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != StringValueType {
		return []bool{}, errors.Errorf("valueType should be bool but get: %s", component.ValueType)
	}
	resp, err = cast.ToBoolSliceE(component.Value)
	return
}

func (ipt *Ipt) GetJsonStrMapValue(componentIndex int) (map[string]interface{}, error) {
	if len(ipt.Components) < componentIndex-1 {
		return map[string]interface{}{}, errors.New("index out of range")
	}
	component := ipt.Components[componentIndex]
	if component.ValueType != JsonValueType {
		return map[string]interface{}{}, errors.Errorf("valueType should be json but get: %s", component.ValueType)
	}
	if !component.AllowMulti {
		return map[string]interface{}{}, nil
	}
	return cast.ToStringMap(component.Value), nil
}

type Ipts []*Ipt

func (iS *Ipts) iptIndexValid(iptIndex int) error {
	if len(*iS) < iptIndex-1 {
		return errors.New("iptIndex out of range")
	}
	return nil
}

func (iS *Ipts) GetIntValue(iptIndex int, componentIndex int) (int, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return 0, err
	}
	return (*iS)[iptIndex].GetIntValue(componentIndex)
}

func (iS *Ipts) GetIntSliceValue(iptIndex int, componentIndex int) ([]int, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []int{0}, err
	}
	return (*iS)[iptIndex].GetIntSliceValue(componentIndex)
}

func (iS *Ipts) GetFloat64Value(iptIndex int, componentIndex int) (float64, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return 0, err
	}
	return (*iS)[iptIndex].GetFloat64Value(componentIndex)
}

func (iS *Ipts) GetFloat64SliceValue(iptIndex int, componentIndex int) ([]float64, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []float64{}, err
	}
	return (*iS)[iptIndex].GetFloat64SliceValue(componentIndex)
}

func (iS *Ipts) GetStringValue(iptIndex int, componentIndex int) (string, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return "", err
	}
	return (*iS)[iptIndex].GetStringValue(componentIndex)
}

func (iS *Ipts) GetStringSliceValue(iptIndex int, componentIndex int) ([]string, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []string{}, err
	}
	return (*iS)[iptIndex].GetStringSliceValue(componentIndex)
}

func (iS *Ipts) GetBoolValue(iptIndex int, componentIndex int) (bool, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return false, err
	}
	return (*iS)[iptIndex].GetBoolValue(componentIndex)
}

func (iS *Ipts) GetBoolSliceValue(iptIndex int, componentIndex int) ([]bool, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return []bool{}, err
	}
	return (*iS)[iptIndex].GetBoolSliceValue(componentIndex)
}

func (iS *Ipts) GetJsonStrMapValue(iptIndex int, componentIndex int) (map[string]interface{}, error) {
	err := iS.iptIndexValid(iptIndex)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return (*iS)[iptIndex].GetJsonStrMapValue(componentIndex)
}
