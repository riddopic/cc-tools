package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/instinct"
)

const (
	defaultConfBarWidth       = 10
	defaultExportFormat       = "yaml"
	evolveCommandConfidence   = 0.7
	evolveAgentAvgConfidence  = 0.75
	evolveMinClusterForAgents = 3
)

func newInstinctCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instinct",
		Short: "Manage learned instincts",
	}
	cmd.AddCommand(
		newInstinctStatusCmd(),
		newInstinctExportCmd(),
		newInstinctImportCmd(),
		newInstinctEvolveCmd(),
	)
	return cmd
}

func newInstinctStatusCmd() *cobra.Command {
	var (
		domain        string
		minConfidence float64
	)

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "List instincts grouped by domain with confidence bars",
		Example: "  cc-tools instinct status --domain testing --min-confidence 0.5",
		RunE: func(_ *cobra.Command, _ []string) error {
			store := newInstinctStore()
			return runInstinctStatus(os.Stdout, store, domain, minConfidence)
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "filter by domain")
	cmd.Flags().Float64Var(&minConfidence, "min-confidence", 0, "minimum confidence threshold")
	return cmd
}

func newInstinctExportCmd() *cobra.Command {
	var (
		output        string
		domain        string
		minConfidence float64
		format        string
	)

	cmd := &cobra.Command{
		Use:     "export",
		Short:   "Export instincts to YAML or JSON",
		Example: "  cc-tools instinct export --format json --output instincts.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			store := newInstinctStore()
			return runInstinctExport(os.Stdout, store, output, domain, minConfidence, format)
		},
	}
	cmd.Flags().StringVar(&output, "output", "", "output file path (default: stdout)")
	cmd.Flags().StringVar(&domain, "domain", "", "filter by domain")
	cmd.Flags().Float64Var(&minConfidence, "min-confidence", 0, "minimum confidence threshold")
	cmd.Flags().StringVar(&format, "format", defaultExportFormat, "output format (yaml or json)")
	return cmd
}

func newInstinctImportCmd() *cobra.Command {
	var (
		dryRun        bool
		force         bool
		minConfidence float64
	)

	cmd := &cobra.Command{
		Use:     "import <source>",
		Short:   "Import instincts from a file",
		Args:    cobra.ExactArgs(1),
		Example: "  cc-tools instinct import instincts.yaml --dry-run",
		RunE: func(_ *cobra.Command, args []string) error {
			store := newInstinctStore()
			return runInstinctImport(os.Stdout, store, args[0], dryRun, force, minConfidence)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be imported without saving")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing instincts")
	cmd.Flags().Float64Var(&minConfidence, "min-confidence", 0, "minimum confidence threshold")
	return cmd
}

func newInstinctEvolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "evolve",
		Short:   "Analyze instinct clusters and suggest skill/command/agent candidates",
		Example: "  cc-tools instinct evolve",
		RunE: func(_ *cobra.Command, _ []string) error {
			store := newInstinctStore()
			cfg := config.GetDefaultConfig()
			return runInstinctEvolve(os.Stdout, store, cfg.Instinct.ClusterThreshold)
		},
	}
}

// newInstinctStore creates a FileStore using configured paths.
func newInstinctStore() *instinct.FileStore {
	cfg := config.GetDefaultConfig()
	personalPath := expandTilde(cfg.Instinct.PersonalPath)
	inheritedPath := expandTilde(cfg.Instinct.InheritedPath)
	return instinct.NewFileStore(personalPath, inheritedPath)
}

// runInstinctStatus lists instincts grouped by domain with confidence bars.
func runInstinctStatus(w io.Writer, store *instinct.FileStore, domain string, minConf float64) error {
	opts := instinct.ListOptions{Domain: domain, MinConfidence: minConf, Source: ""}

	listed, err := store.List(opts)
	if err != nil {
		return fmt.Errorf("list instincts: %w", err)
	}

	if len(listed) == 0 {
		fmt.Fprintln(w, "No instincts found.")
		return nil
	}

	groups := groupByDomain(listed)
	domains := sortedKeys(groups)

	for _, d := range domains {
		printDomainGroup(w, d, groups[d])
	}

	return nil
}

// groupByDomain groups instincts by their domain field.
func groupByDomain(instincts []instinct.Instinct) map[string][]instinct.Instinct {
	groups := make(map[string][]instinct.Instinct)
	for _, inst := range instincts {
		d := inst.Domain
		if d == "" {
			d = "general"
		}
		groups[d] = append(groups[d], inst)
	}
	return groups
}

// sortedKeys returns map keys in sorted order.
func sortedKeys(m map[string][]instinct.Instinct) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// printDomainGroup writes a domain header and its instincts to w.
func printDomainGroup(w io.Writer, domain string, instincts []instinct.Instinct) {
	fmt.Fprintf(w, "\n[%s]\n", domain)
	for _, inst := range instincts {
		bar := confidenceBar(inst.Confidence, defaultConfBarWidth)
		fmt.Fprintf(w, "  %s %.2f  %s\n", bar, inst.Confidence, inst.Trigger)
	}
}

// runInstinctExport exports filtered instincts to a file or stdout.
func runInstinctExport(
	w io.Writer,
	store *instinct.FileStore,
	outputPath, domain string,
	minConf float64,
	format string,
) error {
	opts := instinct.ListOptions{Domain: domain, MinConfidence: minConf, Source: ""}

	instincts, err := store.List(opts)
	if err != nil {
		return fmt.Errorf("list instincts: %w", err)
	}

	if len(instincts) == 0 {
		fmt.Fprintln(w, "No instincts to export.")
		return nil
	}

	dest, err := resolveExportDest(w, outputPath)
	if err != nil {
		return err
	}

	if f, ok := dest.(*os.File); ok {
		defer func() { _ = f.Close() }()
	}

	return writeExport(dest, instincts, format)
}

// resolveExportDest returns the writer for export output.
func resolveExportDest(w io.Writer, outputPath string) (io.Writer, error) {
	if outputPath == "" {
		return w, nil
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create output file: %w", err)
	}

	return f, nil
}

// writeExport writes instincts in the specified format.
func writeExport(w io.Writer, instincts []instinct.Instinct, format string) error {
	switch format {
	case "json":
		return instinct.ExportJSON(w, instincts)
	case "yaml":
		return instinct.ExportYAML(w, instincts)
	default:
		return fmt.Errorf("unsupported format: %s (use yaml or json)", format)
	}
}

// runInstinctImport reads instincts from a file and saves them to the inherited directory.
func runInstinctImport(
	w io.Writer,
	store *instinct.FileStore,
	source string,
	dryRun, force bool,
	minConf float64,
) error {
	parsed, err := readAndParseSource(source)
	if err != nil {
		return err
	}

	if len(parsed) == 0 {
		fmt.Fprintln(w, "No instincts found in source file.")
		return nil
	}

	inherited := newInheritedStore()

	return importInstincts(w, store, inherited, parsed, dryRun, force, minConf)
}

// readAndParseSource reads a file and parses its frontmatter.
func readAndParseSource(source string) ([]instinct.Instinct, error) {
	cleanPath := filepath.Clean(source)
	if strings.Contains(cleanPath, "..") {
		return nil, errors.New("invalid path: directory traversal detected")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("read source file: %w", err)
	}

	parsed, err := instinct.ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse source file: %w", err)
	}

	return parsed, nil
}

// newInheritedStore creates a FileStore that writes to the inherited directory.
func newInheritedStore() *instinct.FileStore {
	cfg := config.GetDefaultConfig()
	inheritedPath := expandTilde(cfg.Instinct.InheritedPath)
	return instinct.NewFileStore(inheritedPath, "")
}

// importInstincts processes parsed instincts and saves eligible ones.
func importInstincts(
	w io.Writer,
	readStore, writeStore *instinct.FileStore,
	parsed []instinct.Instinct,
	dryRun, force bool,
	minConf float64,
) error {
	var imported int

	for _, inst := range parsed {
		action := classifyImport(readStore, inst, force, minConf)
		if action == importSkip {
			continue
		}
		imported++

		label := describeImportAction(action, dryRun)
		fmt.Fprintf(w, "%s %s (%.2f) [%s]\n", label, inst.ID, inst.Confidence, inst.Domain)

		if !dryRun {
			if err := writeStore.Save(inst); err != nil {
				return fmt.Errorf("save instinct %s: %w", inst.ID, err)
			}
		}
	}

	fmt.Fprintf(w, "\n%d instinct(s) %s.\n", imported, importVerb(dryRun))
	return nil
}

type importAction int

const (
	importSkip importAction = iota
	importNew
	importOverwrite
)

// classifyImport determines whether an instinct should be imported.
func classifyImport(store *instinct.FileStore, inst instinct.Instinct, force bool, minConf float64) importAction {
	if minConf > 0 && inst.Confidence < minConf {
		return importSkip
	}

	_, err := store.Get(inst.ID)
	if err == nil && !force {
		return importSkip
	}

	if err == nil {
		return importOverwrite
	}

	return importNew
}

// describeImportAction returns a label for the import action.
func describeImportAction(action importAction, dryRun bool) string {
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}

	switch action {
	case importSkip:
		return prefix + "skip:"
	case importNew:
		return prefix + "import:"
	case importOverwrite:
		return prefix + "overwrite:"
	}

	return prefix + "import:"
}

// importVerb returns the appropriate past-tense verb for import reporting.
func importVerb(dryRun bool) string {
	if dryRun {
		return "would be imported"
	}
	return "imported"
}

// runInstinctEvolve analyzes instinct clusters and suggests candidates.
func runInstinctEvolve(w io.Writer, store *instinct.FileStore, clusterThreshold int) error {
	allInstincts, err := store.List(instinct.ListOptions{Domain: "", MinConfidence: 0, Source: ""})
	if err != nil {
		return fmt.Errorf("list instincts: %w", err)
	}

	if len(allInstincts) < clusterThreshold {
		fmt.Fprintf(w, "Need at least %d instincts to analyze (found %d).\n",
			clusterThreshold, len(allInstincts))
		return nil
	}

	clusters := instinct.ClusterByTrigger(allInstincts, clusterThreshold)
	if len(clusters) == 0 {
		fmt.Fprintln(w, "No clusters found.")
		return nil
	}

	printSkillCandidates(w, clusters)
	printCommandCandidates(w, allInstincts)
	printAgentCandidates(w, clusters)
	return nil
}

// printSkillCandidates prints clusters with 3+ instincts in the same domain.
func printSkillCandidates(w io.Writer, clusters []instinct.Cluster) {
	fmt.Fprintln(w, "\nSkill candidates (3+ related instincts):")

	found := false

	for _, c := range clusters {
		domain := dominantDomain(c.Members)
		if domain == "" {
			continue
		}
		found = true
		fmt.Fprintf(w, "  [%s] %d instincts, keywords: %s\n",
			domain, len(c.Members), strings.Join(c.Keywords, ", "))
	}

	if !found {
		fmt.Fprintln(w, "  (none)")
	}
}

// printCommandCandidates prints high-confidence workflow instincts.
func printCommandCandidates(w io.Writer, allInstincts []instinct.Instinct) {
	fmt.Fprintf(w, "\nCommand candidates (confidence >= %.1f, workflow domain):\n",
		evolveCommandConfidence)

	found := false

	for _, inst := range allInstincts {
		if inst.Confidence >= evolveCommandConfidence && inst.Domain == "workflow" {
			found = true
			fmt.Fprintf(w, "  %.2f  %s\n", inst.Confidence, inst.Trigger)
		}
	}

	if !found {
		fmt.Fprintln(w, "  (none)")
	}
}

// printAgentCandidates prints large clusters with high average confidence.
func printAgentCandidates(w io.Writer, clusters []instinct.Cluster) {
	fmt.Fprintf(w, "\nAgent candidates (%d+ instincts, avg confidence >= %.2f):\n",
		evolveMinClusterForAgents, evolveAgentAvgConfidence)

	found := false

	for _, c := range clusters {
		if len(c.Members) >= evolveMinClusterForAgents && c.AvgConfidence >= evolveAgentAvgConfidence {
			found = true
			fmt.Fprintf(w, "  %d instincts (avg %.2f), keywords: %s\n",
				len(c.Members), c.AvgConfidence, strings.Join(c.Keywords, ", "))
		}
	}

	if !found {
		fmt.Fprintln(w, "  (none)")
	}
}

// dominantDomain returns the most common domain in a set of instincts,
// or empty string if no single domain dominates.
func dominantDomain(members []instinct.Instinct) string {
	counts := make(map[string]int)
	for _, m := range members {
		d := m.Domain
		if d == "" {
			d = "general"
		}
		counts[d]++
	}

	var best string
	var bestCount int

	for d, count := range counts {
		if count > bestCount {
			best = d
			bestCount = count
		}
	}

	return best
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path[1:])
}

// confidenceBar returns a visual confidence indicator.
func confidenceBar(confidence float64, width int) string {
	filled := int(confidence * float64(width))
	empty := width - filled
	return strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", empty)
}
