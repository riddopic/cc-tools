package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/riddopic/cc-tools/internal/config"
	"github.com/riddopic/cc-tools/internal/instinct"
)

const (
	defaultConfBarWidth      = 10
	defaultExportFormat      = "yaml"
	evolveCommandConfidence  = 0.7
	evolveAgentMinCluster    = 3
	evolveAgentAvgConfidence = 0.75
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
			cfg := loadInstinctConfig()
			store := newInstinctStoreFromConfig(cfg)
			return runInstinctStatus(os.Stdout, store, domain, minConfidence, cfg.Instinct.DecayRate)
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
			cfg := loadInstinctConfig()
			store := newInstinctStoreFromConfig(cfg)
			return runInstinctExport(os.Stdout, store, output, domain, minConfidence, format, cfg.Instinct.DecayRate)
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
			cfg := loadInstinctConfig()
			store := newInstinctStoreFromConfig(cfg)
			return runInstinctImport(os.Stdout, store, args[0], dryRun, force, minConfidence, cfg.Instinct.DecayRate)
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
			cfg := loadInstinctConfig()
			store := newInstinctStoreFromConfig(cfg)
			return runInstinctEvolve(os.Stdout, store, cfg.Instinct.ClusterThreshold, cfg.Instinct.DecayRate)
		},
	}
}

// loadInstinctConfig resolves runtime config via the manager, falling back to
// defaults if the config file cannot be loaded.
func loadInstinctConfig() *config.Values {
	mgr := config.NewManager()

	cfg, err := mgr.GetConfig(context.Background())
	if err != nil {
		return config.GetDefaultConfig()
	}

	return cfg
}

// newInstinctStoreFromConfig creates a FileStore using the given config values.
func newInstinctStoreFromConfig(cfg *config.Values) *instinct.FileStore {
	personalPath := expandTilde(cfg.Instinct.PersonalPath)
	inheritedPath := expandTilde(cfg.Instinct.InheritedPath)
	return instinct.NewFileStore(personalPath, inheritedPath)
}

// newInheritedStore creates a FileStore that writes to the inherited directory.
func newInheritedStore() *instinct.FileStore {
	cfg := loadInstinctConfig()
	inheritedPath := expandTilde(cfg.Instinct.InheritedPath)
	return instinct.NewFileStore(inheritedPath, "")
}

// runInstinctStatus lists instincts grouped by domain with confidence bars.
// Decay is applied at display time without mutating stored files.
func runInstinctStatus(w io.Writer, store *instinct.FileStore, domain string, minConf, decayRate float64) error {
	opts := instinct.ListOptions{Domain: domain, MinConfidence: minConf, Source: ""}

	listed, err := store.List(opts)
	if err != nil {
		return fmt.Errorf("list instincts: %w", err)
	}

	if len(listed) == 0 {
		fmt.Fprintln(w, "No instincts found.")
		return nil
	}

	listed = instinct.ApplyDecayToSlice(listed, time.Now(), decayRate)

	groups := instinct.GroupByDomain(listed)
	domains := instinct.SortedKeys(groups)

	for _, d := range domains {
		printDomainGroup(w, d, groups[d])
	}

	return nil
}

// runInstinctExport exports filtered instincts to a file or stdout.
// Decay is applied before export without mutating stored files.
func runInstinctExport(
	w io.Writer,
	store *instinct.FileStore,
	outputPath, domain string,
	minConf float64,
	format string,
	decayRate float64,
) error {
	opts := instinct.ListOptions{Domain: domain, MinConfidence: minConf, Source: ""}

	listed, err := store.List(opts)
	if err != nil {
		return fmt.Errorf("list instincts: %w", err)
	}

	if len(listed) == 0 {
		fmt.Fprintln(w, "No instincts to export.")
		return nil
	}

	listed = instinct.ApplyDecayToSlice(listed, time.Now(), decayRate)

	dest, err := resolveExportDest(w, outputPath)
	if err != nil {
		return err
	}

	if f, ok := dest.(*os.File); ok {
		defer func() { _ = f.Close() }()
	}

	return instinct.Export(dest, listed, format)
}

// runInstinctImport reads instincts from a file and saves them to the inherited
// directory. Decay is applied to parsed instincts before classification and
// the decayed values are persisted on write.
func runInstinctImport(
	w io.Writer,
	store *instinct.FileStore,
	source string,
	dryRun, force bool,
	minConf, decayRate float64,
) error {
	parsed, err := instinct.ReadAndParseSource(source)
	if err != nil {
		return err
	}

	if len(parsed) == 0 {
		fmt.Fprintln(w, "No instincts found in source file.")
		return nil
	}

	parsed = instinct.ApplyDecayToSlice(parsed, time.Now(), decayRate)

	inherited := newInheritedStore()
	opts := instinct.ImportOptions{
		DryRun:        dryRun,
		Force:         force,
		MinConfidence: minConf,
	}

	result, err := instinct.Import(store, inherited, parsed, opts)
	if err != nil {
		return err
	}

	for _, item := range result.Items {
		if item.Action == instinct.ImportSkip {
			continue
		}

		label := item.Action.Label(dryRun)
		fmt.Fprintf(w, "%s %s (%.2f) [%s]\n",
			label, item.Instinct.ID, item.Instinct.Confidence, item.Instinct.Domain)
	}

	fmt.Fprintf(w, "\n%d instinct(s) %s.\n", result.Imported(), result.Verb(dryRun))
	return nil
}

// runInstinctEvolve analyzes instinct clusters and suggests candidates.
// Decay is applied before analysis without mutating stored files.
func runInstinctEvolve(w io.Writer, store *instinct.FileStore, clusterThreshold int, decayRate float64) error {
	allInstincts, err := store.List(instinct.ListOptions{Domain: "", MinConfidence: 0, Source: ""})
	if err != nil {
		return fmt.Errorf("list instincts: %w", err)
	}

	if len(allInstincts) < clusterThreshold {
		fmt.Fprintf(w, "Need at least %d instincts to analyze (found %d).\n",
			clusterThreshold, len(allInstincts))
		return nil
	}

	allInstincts = instinct.ApplyDecayToSlice(allInstincts, time.Now(), decayRate)

	opts := instinct.EvolveOptions{
		ClusterThreshold:   clusterThreshold,
		CommandConfidence:  evolveCommandConfidence,
		CommandDomain:      "workflow",
		AgentMinCluster:    evolveAgentMinCluster,
		AgentAvgConfidence: evolveAgentAvgConfidence,
	}

	result := instinct.Evolve(allInstincts, opts)
	printSkillCandidates(w, result.Skills)
	printCommandCandidates(w, result.Commands, opts.CommandConfidence)
	printAgentCandidates(w, result.Agents, opts.AgentMinCluster, opts.AgentAvgConfidence)
	return nil
}

// printSkillCandidates prints clusters that could become skills.
func printSkillCandidates(w io.Writer, skills []instinct.SkillCandidate) {
	fmt.Fprintln(w, "\nSkill candidates (3+ related instincts):")

	if len(skills) == 0 {
		fmt.Fprintln(w, "  (none)")
		return
	}

	for _, s := range skills {
		fmt.Fprintf(w, "  [%s] %d instincts, keywords: %s\n",
			s.Domain, s.Count, strings.Join(s.Keywords, ", "))
	}
}

// printCommandCandidates prints high-confidence workflow instincts.
func printCommandCandidates(w io.Writer, commands []instinct.CommandCandidate, minConf float64) {
	fmt.Fprintf(w, "\nCommand candidates (confidence >= %.1f, workflow domain):\n", minConf)

	if len(commands) == 0 {
		fmt.Fprintln(w, "  (none)")
		return
	}

	for _, cmd := range commands {
		fmt.Fprintf(w, "  %.2f  %s\n", cmd.Confidence, cmd.Trigger)
	}
}

// printAgentCandidates prints large clusters with high average confidence.
func printAgentCandidates(
	w io.Writer,
	agents []instinct.AgentCandidate,
	minCluster int,
	minAvgConf float64,
) {
	fmt.Fprintf(w, "\nAgent candidates (%d+ instincts, avg confidence >= %.2f):\n",
		minCluster, minAvgConf)

	if len(agents) == 0 {
		fmt.Fprintln(w, "  (none)")
		return
	}

	for _, a := range agents {
		fmt.Fprintf(w, "  %d instincts (avg %.2f), keywords: %s\n",
			a.Count, a.AvgConfidence, strings.Join(a.Keywords, ", "))
	}
}

// printDomainGroup writes a domain header and its instincts to w.
func printDomainGroup(w io.Writer, domain string, instincts []instinct.Instinct) {
	fmt.Fprintf(w, "\n[%s]\n", domain)
	for _, inst := range instincts {
		bar := confidenceBar(inst.Confidence, defaultConfBarWidth)
		fmt.Fprintf(w, "  %s %.2f  %s\n", bar, inst.Confidence, inst.Trigger)
	}
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
