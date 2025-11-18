package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type swaggerDoc struct {
	Paths map[string]map[string]operation `json:"paths"`
}

type operation struct {
	Summary     string                   `json:"summary"`
	Description string                   `json:"description"`
	Tags        []string                 `json:"tags"`
	Security    []map[string]interface{} `json:"security"`
}

type row struct {
	Service string
	Path    string
	Method  string
	Summary string
	Tags    string
	Auth    string
	Notes   string
}

func (r row) key() string {
	return r.Service + "|" + r.Method + "|" + r.Path
}

func main() {
	output := flag.String("output", "test/reports/swagger_matrix.csv", "output csv path")
	flag.Parse()

	inputs := []struct {
		service string
		path    string
	}{
		{service: "admin", path: "docs/admin/swagger.json"},
		{service: "game", path: "docs/game/swagger.json"},
	}

	existingNotes := readNotes(*output)

	var rows []row
	for _, in := range inputs {
		doc, err := loadSwagger(in.path)
		if err != nil {
			panic(fmt.Sprintf("load %s: %v", in.path, err))
		}
		rows = append(rows, flatten(in.service, doc)...)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Service == rows[j].Service {
			return rows[i].Path+rows[i].Method < rows[j].Path+rows[j].Method
		}
		return rows[i].Service < rows[j].Service
	})

	for i := range rows {
		if note, ok := existingNotes[rows[i].key()]; ok {
			rows[i].Notes = note
		}
	}

	if err := writeCSV(*output, rows); err != nil {
		panic(err)
	}

	fmt.Printf("exported %d rows to %s\n", len(rows), *output)
}

func loadSwagger(path string) (*swaggerDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc swaggerDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func flatten(service string, doc *swaggerDoc) []row {
	var out []row
	for p, methods := range doc.Paths {
		for method, op := range methods {
			out = append(out, row{
				Service: service,
				Path:    p,
				Method:  strings.ToUpper(method),
				Summary: op.Summary,
				Tags:    strings.Join(op.Tags, ";"),
				Auth:    authRequired(op),
			})
		}
	}
	return out
}

func authRequired(op operation) string {
	if len(op.Security) == 0 {
		return "false"
	}
	return "true"
}

func writeCSV(path string, rows []row) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	headers := []string{"service", "method", "path", "summary", "tags", "auth_required", "notes"}
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{r.Service, r.Method, r.Path, r.Summary, r.Tags, r.Auth, r.Notes}); err != nil {
			return err
		}
	}
	return w.Error()
}

func readNotes(path string) map[string]string {
	notes := make(map[string]string)
	f, err := os.Open(path)
	if err != nil {
		return notes
	}
	defer f.Close()

	r := csv.NewReader(f)
	headers, err := r.Read()
	if err != nil {
		return notes
	}
	indices := map[string]int{}
	for i, h := range headers {
		indices[h] = i
	}
	noteIdx, ok := indices["notes"]
	if !ok {
		return notes
	}
	serviceIdx := indices["service"]
	methodIdx := indices["method"]
	pathIdx := indices["path"]
	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		note := record[noteIdx]
		if note == "" {
			continue
		}
		key := record[serviceIdx] + "|" + record[methodIdx] + "|" + record[pathIdx]
		notes[key] = note
	}
	return notes
}
