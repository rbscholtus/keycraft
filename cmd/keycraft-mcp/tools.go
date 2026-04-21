package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/rbscholtus/keycraft/internal/api"
	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// toolDef is the server-internal representation of a tool — its name, an
// input schema for clients, and the handler that runs when the tool is
// called.
type toolDef struct {
	Name        string
	Description string
	InputSchema map[string]any
	handler     func(ctx context.Context, s *server, raw json.RawMessage) (any, error)
}

// findTool returns the named tool or nil.
func (s *server) findTool(name string) *toolDef {
	for i := range s.tools {
		if s.tools[i].Name == name {
			return &s.tools[i]
		}
	}
	return nil
}

// toolListPayload formats the tool registry for the tools/list response.
func (s *server) toolListPayload() []map[string]any {
	out := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		out = append(out, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return out
}

// buildTools is the static tool registry. Each handler decodes its DTO and
// delegates to api.Service.
func (s *server) buildTools() []toolDef {
	return []toolDef{
		{
			Name:        "list_corpora",
			Description: "List available text corpora in the data-dir.",
			InputSchema: emptyObjectSchema(),
			handler: func(_ context.Context, s *server, _ json.RawMessage) (any, error) {
				names, err := s.svc.ListCorpora()
				if err != nil {
					return nil, err
				}
				return map[string]any{"corpora": names}, nil
			},
		},
		{
			Name:        "describe_corpus",
			Description: "Load (or reuse cached) a corpus and return its summary statistics.",
			InputSchema: objectSchema(map[string]any{
				"name":     stringSchema("Corpus filename, e.g. 'monkeyracer.txt'."),
				"coverage": numberSchema("Word-coverage percentage (0.1-100). Defaults to 98."),
			}, []string{"name"}),
			handler: func(_ context.Context, s *server, raw json.RawMessage) (any, error) {
				var ref api.CorpusRef
				if err := decodeArgs(raw, &ref); err != nil {
					return nil, err
				}
				return s.svc.DescribeCorpus(s.sessionID, ref)
			},
		},
		{
			Name:        "list_layouts",
			Description: "List .klf layout files in the data-dir.",
			InputSchema: emptyObjectSchema(),
			handler: func(_ context.Context, s *server, _ json.RawMessage) (any, error) {
				names, err := s.svc.ListLayouts()
				if err != nil {
					return nil, err
				}
				return map[string]any{"layouts": names}, nil
			},
		},
		{
			Name:        "load_layout",
			Description: "Load a single layout by name and return its grid.",
			InputSchema: objectSchema(map[string]any{
				"name": stringSchema("Layout name (with or without .klf suffix)."),
			}, []string{"name"}),
			handler: func(_ context.Context, s *server, raw json.RawMessage) (any, error) {
				var args struct{ Name string }
				if err := decodeArgs(raw, &args); err != nil {
					return nil, err
				}
				return s.svc.GetLayout(args.Name)
			},
		},
		{
			Name:        "flip_layout",
			Description: "Flip a layout horizontally and return the result. Does NOT save to disk; use save_layout to persist.",
			InputSchema: objectSchema(map[string]any{
				"name": stringSchema("Layout name to flip."),
			}, []string{"name"}),
			handler: func(_ context.Context, s *server, raw json.RawMessage) (any, error) {
				var args struct{ Name string }
				if err := decodeArgs(raw, &args); err != nil {
					return nil, err
				}
				return s.svc.FlipLayout(args.Name)
			},
		},
		{
			Name:        "save_layout",
			Description: "Persist the named source layout under a new filename in the data-dir's layouts/ folder. Useful after flip_layout.",
			InputSchema: objectSchema(map[string]any{
				"source": stringSchema("Source layout name to save under a new name."),
				"dest":   stringSchema("Destination .klf filename (without path; saved into data-dir/layouts/)."),
				"flip":   boolSchema("If true, also flip the source horizontally before saving."),
			}, []string{"source", "dest"}),
			handler: func(_ context.Context, s *server, raw json.RawMessage) (any, error) {
				var args struct {
					Source, Dest string
					Flip         bool
				}
				if err := decodeArgs(raw, &args); err != nil {
					return nil, err
				}
				return saveLayout(s, args.Source, args.Dest, args.Flip)
			},
		},
		{
			Name:        "analyse",
			Description: "Compute detailed metrics for one or more layouts against a corpus.",
			InputSchema: analyseSchema(),
			handler: func(_ context.Context, s *server, raw json.RawMessage) (any, error) {
				var req api.AnalyseRequest
				if err := decodeArgs(raw, &req); err != nil {
					return nil, err
				}
				return s.svc.Analyse(s.sessionID, req)
			},
		},
		{
			Name:        "rank",
			Description: "Rank layouts by weighted score against a corpus.",
			InputSchema: rankSchema(),
			handler: func(_ context.Context, s *server, raw json.RawMessage) (any, error) {
				var req api.RankRequest
				if err := decodeArgs(raw, &req); err != nil {
					return nil, err
				}
				return s.svc.Rank(s.sessionID, req)
			},
		},
		{
			Name:        "optimise_fast",
			Description: "Run BLS local-search only (no perturbation) on a layout. Fast: typically seconds.",
			InputSchema: optimiseFastSchema(),
			handler: func(ctx context.Context, s *server, raw json.RawMessage) (any, error) {
				var req api.OptimiseFastRequest
				if err := decodeArgs(raw, &req); err != nil {
					return nil, err
				}
				return s.svc.OptimiseFast(ctx, s.sessionID, req)
			},
		},
		{
			Name:        "generate_fast",
			Description: "Generate layouts from a .gen config without running optimisation. Persists results under data-dir/layouts/.",
			InputSchema: generateFastSchema(),
			handler: func(ctx context.Context, s *server, raw json.RawMessage) (any, error) {
				var req api.GenerateFastRequest
				if err := decodeArgs(raw, &req); err != nil {
					return nil, err
				}
				return s.svc.GenerateFast(ctx, s.sessionID, req)
			},
		},
	}
}

// saveLayout loads `source`, optionally flips it, and persists under `dest`
// in data-dir/layouts/. Refuses absolute paths or any traversal outside the
// layouts directory.
func saveLayout(s *server, source, dest string, flip bool) (any, error) {
	if dest == "" {
		return nil, fmt.Errorf("dest is required")
	}
	if filepath.IsAbs(dest) || filepath.Clean(dest) != filepath.Base(filepath.Clean(dest)) {
		return nil, fmt.Errorf("dest must be a plain filename, no slashes")
	}

	// Confirm source exists (via the service so behaviour matches other tools)
	// then load the actual SplitLayout for mutation.
	if _, err := s.svc.GetLayout(source); err != nil {
		return nil, err
	}
	cfg := s.svc.Manager().Config()
	srcPath := filepath.Join(cfg.LayoutDir(), ensureKlf(source))
	layout, err := kc.NewLayoutFromFile(trimKlf(filepath.Base(srcPath)), srcPath)
	if err != nil {
		return nil, err
	}
	if flip {
		layout.FlipHorizontal()
	}
	layout.Name = trimKlf(dest)

	outPath := filepath.Join(cfg.LayoutDir(), ensureKlf(dest))
	if err := layout.SaveToFile(outPath); err != nil {
		return nil, err
	}
	return map[string]any{"saved": outPath, "name": layout.Name}, nil
}

func ensureKlf(name string) string {
	if filepath.Ext(name) == ".klf" {
		return name
	}
	return name + ".klf"
}

func trimKlf(name string) string {
	if ext := filepath.Ext(name); ext == ".klf" {
		return name[:len(name)-len(ext)]
	}
	return name
}

// decodeArgs unmarshals MCP tool arguments into the destination, treating an
// empty raw payload as an empty object.
func decodeArgs(raw json.RawMessage, dst any) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, dst)
}

// ---------- JSON Schema helpers ----------

func emptyObjectSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func objectSchema(props map[string]any, required []string) map[string]any {
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func stringSchema(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func numberSchema(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}

func boolSchema(desc string) map[string]any {
	return map[string]any{"type": "boolean", "description": desc}
}

func corpusRefSchema() map[string]any {
	return objectSchema(map[string]any{
		"name":     stringSchema("Corpus filename, e.g. 'monkeyracer.txt'."),
		"coverage": numberSchema("Word-coverage percentage. Defaults to 98."),
	}, []string{"name"})
}

func analyseSchema() map[string]any {
	return objectSchema(map[string]any{
		"layouts": map[string]any{
			"type":        "array",
			"items":       stringSchema("Layout name."),
			"description": "Layout names to analyse.",
		},
		"corpus": corpusRefSchema(),
	}, []string{"layouts", "corpus"})
}

func rankSchema() map[string]any {
	return objectSchema(map[string]any{
		"layouts": map[string]any{
			"type":  "array",
			"items": stringSchema("Layout name."),
		},
		"corpus": corpusRefSchema(),
	}, []string{"layouts", "corpus"})
}

func optimiseFastSchema() map[string]any {
	return objectSchema(map[string]any{
		"layout": stringSchema("Layout name to optimise."),
		"corpus": corpusRefSchema(),
		"pins":   stringSchema("Characters to pin (e.g. 'aeiouy')."),
		"free":   stringSchema("Characters free to move (mutually exclusive with pins)."),
		"seed":   numberSchema("Optional random seed."),
	}, []string{"layout", "corpus"})
}

func generateFastSchema() map[string]any {
	return objectSchema(map[string]any{
		"config":      stringSchema("Generation config filename (relative to config/)."),
		"max_layouts": numberSchema("Cap on layouts produced (0 = all). Defaults to 5000."),
		"seed":        numberSchema("Random seed (0 = timestamp)."),
	}, []string{"config"})
}

