package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// convert comma separated list into an array, trimming spaces and ignoring empty values
func commaSeparatedListToSlice(value string) (result []string) {
	for _, val := range strings.Split(value, ",") {
		val = strings.TrimSpace(val)
		if val != "" {
			result = append(result, val)
		}
	}
	return result
}

func commandHelper(cmd *cobra.Command, args []string, mainFunction func(logger *zap.Logger, cmd *cobra.Command, args []string) error) {
	config := zap.NewDevelopmentConfig()
	config.DisableCaller = true
	config.DisableStacktrace = true
	logger, err := config.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing command: %v\n", err)
		os.Exit(1)
	}

	defer logger.Sync()

	err = mainFunction(logger, cmd, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
}
