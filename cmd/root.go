package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jesselang/dox/pkg"
)

var dryRun bool
var cfgFile string
var repoRoot string
var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dox",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		files, err := dox.FindAll(repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}

		root, err := dox.Publish("", "", repoRoot, dryRun) // generated root page
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		if verbose {
			fmt.Printf("root published to %s\n", root)
		}

		for _, v := range files {
			_, err := dox.Publish(v, root, repoRoot, dryRun)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}
		}

		// ham-fisted approach to ensuring that all relative links are linked
		// correctly, first pass above cannot guarantee that all pages were
		// already created when executed
		for _, v := range files {
			id, err := dox.Publish(v, root, repoRoot, dryRun)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}
			if verbose {
				fmt.Printf("%s published to %s\n", v, id)
			}
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repoRoot, err = dox.FindRepoRoot(path)
	if err != nil {
		panic(err)
	}

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dox.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Dry-run mode")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Be verbose")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".dox")   // name of config file (without extension)
	viper.AddConfigPath(repoRoot) // adding repo directory as first search path
	viper.SetEnvPrefix("dox")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
