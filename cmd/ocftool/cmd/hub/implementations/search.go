package implementations

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"github.com/hokaccha/go-prettyjson"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type searchOptions struct {
	interfacePath     string
	output            string
	interfaceRevision string
}

func NewSearch() *cobra.Command {
	var opts searchOptions

	search := &cobra.Command{
		Use:   "search",
		Short: "Lists the currently available Implementations on the Hub server",
		Example: heredoc.WithCLIName(`
			# Show all implementations in table format
			<cli> hub implementations search cap.interface.database.postgresql.install
			
			# Show all implementations in YAML format			
			<cli> hub implementations search cap.interface.database.postgresql.install -oyaml
		`, ocftool.CLIName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.interfacePath = args[0]
			return searchImpl(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := search.Flags()

	flags.StringVar(&opts.interfaceRevision, "interface-revision", "", "Specific interface revision")
	flags.StringVarP(&opts.output, "output", "o", "table", "Output format. One of:\njson | yaml | table")

	return search
}

func searchImpl(ctx context.Context, opts searchOptions, w io.Writer) error {
	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	interfaces, err := cli.ListImplementationRevisionsForInterface(ctx, gqlpublicapi.InterfaceReference{
		Path:     opts.interfacePath,
		Revision: opts.interfaceRevision,
	})
	if err != nil {
		return err
	}

	printInterfaces, err := selectPrinter(opts.output)
	if err != nil {
		return err
	}

	return printInterfaces(interfaces, w)
}

// TODO: all funcs should be extracted to `printers` package and return Printer Interface

type printer func(in []gqlpublicapi.ImplementationRevision, w io.Writer) error

func selectPrinter(format string) (printer, error) {
	switch format {
	case "json":
		return printJSON, nil
	case "yaml":
		return printYAML, nil
	case "table":
		return printTable, nil
	}

	return nil, fmt.Errorf("unknow output format %q", format)
}

func printJSON(in []gqlpublicapi.ImplementationRevision, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(in []gqlpublicapi.ImplementationRevision, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func printTable(in []gqlpublicapi.ImplementationRevision, w io.Writer) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"PATH", "LATEST REVISION", "ATTRIBUTES"})
	table.SetAutoWrapText(true)
	table.SetColumnSeparator(" ")
	table.SetBorder(false)
	table.SetRowLine(true)

	var data [][]string
	for _, item := range in {
		data = append(data, []string{
			item.Metadata.Path,
			item.Revision,
			attrNames(item.Metadata.Attributes),
		})
	}
	table.AppendBulk(data)

	table.Render()

	return nil
}

func attrNames(attrs []*gqlpublicapi.AttributeRevision) string {
	var out []string
	for _, a := range attrs {
		if a == nil || a.Metadata == nil {
			continue
		}
		out = append(out, a.Metadata.Path)
	}

	return strings.Join(out, "\n")
}