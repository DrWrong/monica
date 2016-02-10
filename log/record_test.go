package log

import (
	"testing"
	"text/template"
)

func TestRecord(t *testing.T) {
	const formatter = `
{{.Time.String }}  {{.Level.String }} {{.FileName }} {{.FuncName}} {{ .LineNo}} {{ .Message }}
`
	record := NewRecord(DebugLevel, "this is a test")
	parsedTemplate := template.Must(template.New("logtemplate").Parse(formatter))

	out, err := record.Bytes(parsedTemplate)
	t.Logf("%s", out)
	if err != nil {
		t.Error(err)
	}
}
