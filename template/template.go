package template

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"

	"github.com/google/uuid"
)

// Template stores all the templates loaded
type Template struct {
	dir string
	tpl *template.Template
}

// NewTemplate loads and configures a new template from dir
func NewTemplate(dir string, name string) *Template {
	var t *template.Template
	var tmplFuncMap = template.FuncMap{
		"uuid":      uuid.New,
		"randFloat": rand.Float32,
		"randInt":   rand.Int,
		"toString": func(s io.Reader) (string, error) {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(s)
			return buf.String(), err
		},
		"toJSON": func(obj interface{}) (io.Reader, error) {
			buf := new(bytes.Buffer)
			encoder := json.NewEncoder(buf)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("", "  ")
			err := encoder.Encode(obj)
			return buf, err
		},
		"file": func(filename string) (io.ReadCloser, error) {
			return os.Open(filename)
		},
		"tpl": func(n string) (io.Reader, error) {
			buf := new(bytes.Buffer)
			if err := t.ExecuteTemplate(buf, n, 0); err != nil {
				return buf, err
			}
			return buf, nil
		},
	}

	t = template.Must(
		template.New(filepath.Base(dir)).
			Funcs(tmplFuncMap).
			Option("missingkey=default").
			Delims("{{", "}}").
			ParseGlob(filepath.Join(dir, "*")),
	).Lookup(name)

	// TODO: check t == nil

	return &Template{
		dir: dir,
		tpl: t,
	}
}

func (t *Template) String() string {
	return t.tpl.DefinedTemplates()
}
