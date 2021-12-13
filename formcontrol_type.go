package bloc_client

type FormControlType string

const (
	InputFormControl    FormControlType = "input"
	SelectFormControl   FormControlType = "select"
	RadioFormControl    FormControlType = "radio"
	TextAreaFormControl FormControlType = "textarea"
	JsonFormControl     FormControlType = "json"
)

type SelectOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}
