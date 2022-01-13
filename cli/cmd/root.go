/*
Copyright © 2022 Nicolas MASSE

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	exporter "github.com/nmasse-itix/nftables-counters-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nftables-exporter",
	Short: "Export nftables counters as open metrics format",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		nftc, err := exporter.NewNftablesCounterCollector()
		if err != nil {
			log.Fatalln(err)
		}
		registry := prometheus.NewRegistry()
		registry.MustRegister(nftc)

		var opts promhttp.HandlerOpts
		opts.ErrorLog = log.Default()
		opts.ErrorHandling = promhttp.HTTPErrorOnError
		opts.MaxRequestsInFlight = viper.GetInt("MaxRequestsInFlight")
		opts.Timeout = viper.GetDuration("Timeout")
		opts.EnableOpenMetrics = true
		http.Handle("/metrics", promhttp.HandlerFor(registry, opts))
		listenAddr := viper.GetString("ListenAddr")
		log.Printf("Listening on %s...", listenAddr)
		log.Fatal(http.ListenAndServe(listenAddr, nil))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	log.SetFlags(0)
	viper.SetDefault("ListenAddr", ":9923")
	viper.SetDefault("MaxRequestsInFlight", 5)
	viper.SetDefault("Timeout", "5s")
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/nftables-exporter.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("nftables-exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
