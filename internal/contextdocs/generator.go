// SPDX-License-Identifier: AGPL-3.0-or-later
package contextdocs

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bartekus/cortex/internal/projection"
	"github.com/bartekus/cortex/internal/xray"
)

// Generator projects XRAY index to documentation.
type Generator struct {
	Index  *xray.Index
	OutDir string
}

// Generate renders all context documentation pages.
func (g *Generator) Generate() error {
	if err := g.renderIndex(); err != nil {
		return fmt.Errorf("rendering index.md: %w", err)
	}
	if err := g.renderFiles(); err != nil {
		return fmt.Errorf("rendering files.md: %w", err)
	}
	if err := g.renderModules(); err != nil {
		return fmt.Errorf("rendering modules.md: %w", err)
	}
	return nil
}

func (g *Generator) renderIndex() error {
	var b strings.Builder

	b.WriteString(projection.RenderHeader(1, "Context Index"))

	// Summary Stats
	b.WriteString(projection.RenderHeader(2, "Summary"))
	b.WriteString(fmt.Sprintf("- **Root**: `%s`\n", g.Index.Root))
	b.WriteString(fmt.Sprintf("- **Target**: `%s`\n", g.Index.Target))
	b.WriteString(fmt.Sprintf("- **Digest**: `%s`\n", g.Index.Digest))
	b.WriteString(fmt.Sprintf("- **Files**: %d\n", g.Index.Stats.FileCount))
	b.WriteString(fmt.Sprintf("- **Total Size**: %d bytes\n", g.Index.Stats.TotalSize))
	b.WriteString("\n")

	// Languages
	b.WriteString(projection.RenderHeader(2, "Languages"))
	langKeys := projection.SortedKeys(g.Index.Languages)
	langRows := make([][]string, 0, len(langKeys))
	for _, k := range langKeys {
		count := g.Index.Languages[k]
		langRows = append(langRows, []string{k, strconv.Itoa(count)})
	}
	b.WriteString(projection.RenderTable([]string{"Language", "Files"}, langRows))
	b.WriteString("\n")

	// Top Directories
	b.WriteString(projection.RenderHeader(2, "Top Directories"))
	topDirKeys := projection.SortedKeys(g.Index.TopDirs)
	topDirRows := make([][]string, 0, len(topDirKeys))
	for _, k := range topDirKeys {
		count := g.Index.TopDirs[k]
		topDirRows = append(topDirRows, []string{k, strconv.Itoa(count)})
	}
	b.WriteString(projection.RenderTable([]string{"Directory", "Files"}, topDirRows))
	// b.WriteString("\n")

	return projection.AtomicWrite(filepath.Join(g.OutDir, "index.md"), []byte(b.String()))
}

func (g *Generator) renderFiles() error {
	var b strings.Builder

	b.WriteString(projection.RenderHeader(1, "File Inventory"))

	// Files are already sorted in XRAY index, but we rely on that.
	// Table columns: Path, Size, Language, Complexity

	rows := make([][]string, 0, len(g.Index.Files))
	for _, f := range g.Index.Files {
		// Complexity is currently a placeholder 0 in XRAY.
		// We show it but label it as such in the header or just value.
		// User instruction: Omit or label as placeholder.
		// Decision: Omit for now to avoid confusion, or include as "0".
		// User plan said: "Complexity: Omit or label as 'placeholder'".
		// Let's omit it from the table for cleanliness until it is real.

		rows = append(rows, []string{
			f.Path,
			strconv.FormatInt(f.Size, 10),
			f.Lang,
			strconv.Itoa(f.LOC),
		})
	}

	b.WriteString(projection.RenderTable([]string{"Path", "Size", "Language", "LOC"}, rows))

	return projection.AtomicWrite(filepath.Join(g.OutDir, "files.md"), []byte(b.String()))
}

func (g *Generator) renderModules() error {
	var b strings.Builder

	b.WriteString(projection.RenderHeader(1, "Module Files"))
	b.WriteString("Key configuration files defining modules or dependencies.\n\n")

	// moduleFiles is sorted by XRAY.
	b.WriteString(projection.RenderList(g.Index.ModuleFiles))

	return projection.AtomicWrite(filepath.Join(g.OutDir, "modules.md"), []byte(b.String()))
}
