package main

import (
	"context"
	"fmt"
	"os"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

var (
	deleteCmd = &cobra.Command{
		Short: "Delete metric descriptors.",
		Long:  "Delete unused metric descriptors from stackdriver.",
		Run:   del,
	}

	// flags
	dryRun         bool
	project        string
	matchSubstring string

	metricClient *monitoring.MetricClient
)

func init() {
	deleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "toggle dry run on and off.")
	deleteCmd.Flags().StringVar(&project, "project", "", "GCP project ID")
	deleteCmd.Flags().StringVar(&matchSubstring, "match-substring", "", "match substring in metrics descriptor")

	deleteCmd.MarkFlagRequired("project")
	deleteCmd.MarkFlagRequired("match-substring")
}

func main() {
	var err error
	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		fmt.Fprintln(os.Stderr, "missing required env var GOOGLE_APPLICATION_CREDENTIALS")
		return
	}

	metricClient, err = monitoring.NewMetricClient(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating metric client: %v\n", err)
		return
	}
	defer metricClient.Close()

	deleteCmd.Execute()
}

func del(cmd *cobra.Command, args []string) {
	listReq := &monitoringpb.ListMetricDescriptorsRequest{
		Name:   fmt.Sprintf("projects/%s", project),
		Filter: fmt.Sprintf("metric.type = has_substring(\"%s\")", matchSubstring),
	}
	it := metricClient.ListMetricDescriptors(context.Background(), listReq)
	for {
		m, err := it.Next()
		if err == iterator.Done {
			fmt.Fprintf(os.Stdout, "done: %v\n", err)
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			break
		}

		fmt.Println(m.GetName())
		if !dryRun {
			// delete
			delReq := &monitoringpb.DeleteMetricDescriptorRequest{
				Name: m.GetName(),
			}
			err := metricClient.DeleteMetricDescriptor(context.Background(), delReq)
			if err != nil {
				// error deleting descriptor
				fmt.Fprintf(os.Stderr, "deleting %s: %v\n", m.GetName(), err)
			}
		}
	}
}
