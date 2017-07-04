// Copyright © 2017 Pradip Patil
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"encoding/json"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pradippatil/gist/conf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var files []string
var description string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gist",
	Short: "upload code to https://gist.github.com",
	Long: `gist is a commandline app that you can use from your terminal
to upload content to https://gist.github.com/`,
	Run: upload,
}

func upload(cmd *cobra.Command, args []string) {
	// If we don't have any files to upload, show command usage and exit
	// TODO: check if there's any smarter way to do this.
	if len(files) == 0 {
		cmd.Usage()
		os.Exit(1)
	}

	gist := conf.Gist{}

	gist.Decription = description
	gist.Public = true
	gist.Files = make(map[string]*conf.File)

	// TODO: Add check for file sizes and check if multipart upload is available
	// Refereces:
	// 		https://github.com/sclevine/cflocal/blob/49495238fad2959061bef7a23c6b28da8734f838/remote/droplet.go#L21-201
	//		https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
	for _, f := range files {
		gistFile := conf.File{}
		fs, err := os.Stat(f)
		if err != nil {
			log.Fatal(err)
		}
		fc, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		gistFile.Name = fs.Name()
		gistFile.Content = string(fc)
		gist.Files[fs.Name()] = &gistFile

	}

	// TODO: Add check for below naming constraints
	// Note: Don't name your files "gistfile" with a numerical suffix.
	// This is the format of the automatic naming scheme that Gist uses internally.
	// Check: https://developer.github.com/v3/gists/#create-a-gist

	b, err := json.Marshal(gist)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", conf.GistAPIURL, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == conf.StatusCreated {
		fmt.Println("gist created successfully")
	} else {
		log.Fatal(fmt.Errorf("%s: %v", resp.Status, resp.Header))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var g conf.Gist
	if err := json.Unmarshal(body, &g); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("access it using url: %s\n", g.HTMLURL)

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gist.json)")
	RootCmd.Flags().StringSliceVarP(&files, "files", "f", []string{}, "files to upload")
	RootCmd.Flags().StringVarP(&description, "desc", "d", "my gist", "gist description")
	RootCmd.Flags().SortFlags = false
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		// Search config in home directory with name ".gist" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gist")
	}

	viper.SetEnvPrefix("GIST") // segregate env variables with prefix GIST_
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}

	var logCfg conf.Log
	// FIXME: Overriding log config using env variables like GIST_LOG_LEVEL doesn't work
	if err := viper.UnmarshalKey("log", &logCfg); err != nil {
		log.Fatal(err)
	}

	log, err := conf.InitLogger(&logCfg)
	if err != nil {
		log.Fatal(err)
	}

}
