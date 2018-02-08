package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/victoru/param_api/pkg/ssm"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	rootCmd.Execute()
}

const DEFAULT_ENVIRONMENT = "dev"

var currentEnvironment = DEFAULT_ENVIRONMENT

var rootCmd = &cobra.Command{
	Use:   "paramcli",
	Short: "Load files and environment variables from SSM Parameter Store.",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Takes JSON file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		currentEnvironment = DEFAULT_ENVIRONMENT
		if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
			currentEnvironment = env
		}

		b, err := ioutil.ReadFile(args[0])
		if err != nil {
			log.Fatalf("got error reading file: %s", err)
		}
		m := decodeJSON(b)
		var params []string
		for _, v := range m {
			params = append(params, fmt.Sprintf("/%s/%s", currentEnvironment, v))
		}

		region := os.Getenv("AWS_REGION")
		ssmclient := ssm.NewClient(region)
		o, err := ssmclient.ParamList(params...)
		if err != nil {
			log.Fatal(err)
		}
		data := map[string]string{}
		for k, v := range m {
			for _, par := range o.Parameters {
				if fmt.Sprintf("/%s/%s", currentEnvironment, v) == *par.Name {
					// base64 encode
					data[k] = base64.StdEncoding.EncodeToString([]byte(*par.Value))
				}
			}
		}
		// parse final data and export env vars or create files
		var stmtLines []string
		for k, v := range data {
			if strings.Contains(k, "/") {
				// looks like a file
				if err := createFileFromParam(k, v); err != nil {
					log.Fatalf("error creating file: %s\n", err)
				}
			} else {
				fmt.Printf("export %s=%s\n", k, v)
				stmtLines = append(stmtLines, fmt.Sprintf(`export %s=%s`, k, v))
			}
		}

		log.Println("----")
		//fmt.Println(strings.Join(stmtLines, "\n"))
	},
}

func createFileFromParam(filePath, b64content string) error {

	contents, err := base64.StdEncoding.DecodeString(b64content)
	if err != nil {
		log.Fatalf("got error base64 decoding file contents!\n")
		return err
	}

	filePath = fmt.Sprintf("/tmp/%s", filePath)

	dir := path.Dir(filePath)
	log.Printf("creating directory %s...", dir)
	if err := os.MkdirAll(dir, 0600); err != nil {
		log.Println("unable to create directory")
		return err
	}

	log.Printf("writing to %s...", filePath)
	if err := ioutil.WriteFile(filePath, []byte(contents), 0600); err != nil {
		return err
	}
	return nil
}

func decodeJSON(b []byte) map[string]string {
	decoder := json.NewDecoder(bytes.NewBuffer(b))
	var p map[string]string
	err := decoder.Decode(&p)
	if err != nil {
		log.Fatalf("encountered issue decoding JSON file; %s", err.Error())
	}
	return p
}
