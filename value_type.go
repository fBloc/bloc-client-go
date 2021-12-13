package bloc_client

// ValueType 此控件输入值的类型
type ValueType string

const (
	IntValueType    ValueType = "int"
	FloatValueType  ValueType = "float"
	StringValueType ValueType = "string"
	BoolValueType   ValueType = "bool"
	JsonValueType   ValueType = "json"
)
