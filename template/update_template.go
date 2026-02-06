package template

var updateSchema = map[string]string{
	"Code":  "string",
	"Type":  "string",
	"Value": "string",
}

// UpdateTemplate is a struct that handles interactions with an Insert JSON template
type UpdateTemplate struct {
	data map[string]map[string]map[string]any
}

func (t *UpdateTemplate) TemplateFrom(path string) error {

	return nil
}

func (t *UpdateTemplate) validateTemplate(data map[string]map[string]map[string]any) error {

	return nil
}
