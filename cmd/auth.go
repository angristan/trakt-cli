package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/angristan/trakt-cli/api"
	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

type Credentials struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	AccessToken  string `yaml:"access-token"`
}

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with trakt.tv",
	Long:  "You will need to go to https://trakt.tv/oauth/applications/new to get a access id and secret",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewAPIClient()

		resp, err := client.AuthDeviceCode(&api.AuthDeviceCodeReq{
			ClientID: cmd.Flag("client-id").Value.String(),
		})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get device code1")
			return
		}

		fmt.Printf("Please go to %s and enter the following code: %s\n", resp.VerificationURL, resp.UserCode)

		s := spinner.New(spinner.CharSets[2], 100*time.Millisecond)
		s.Start()
		s.Prefix = "Waiting for authorisation... "

		for {
			tokenResp, err := client.AuthDeviceToken(&api.AuthDeviceTokenReq{
				Code:         resp.DeviceCode,
				ClientID:     cmd.Flag("client-id").Value.String(),
				ClientSecret: cmd.Flag("client-secret").Value.String(),
			})
			if err != nil {
				logrus.WithError(err).Fatal("Failed to get device code")
				return
			}
			if len(tokenResp.AccessToken) == 0 {
				time.Sleep(time.Duration(resp.Interval) * time.Second)
			} else {
				creds := Credentials{
					ClientID:     cmd.Flag("client-id").Value.String(),
					ClientSecret: cmd.Flag("client-secret").Value.String(),
					AccessToken:  tokenResp.AccessToken,
				}

				yamlData, err := yaml.Marshal(&creds)
				if err != nil {
					fmt.Printf("Error while Marshaling. %v", err)
				}

				// write to ~/.trakt.yaml
				homeDir, err := os.UserHomeDir()
				if err != nil {
					log.Fatal(err)
				}
				err = ioutil.WriteFile(homeDir+"/.trakt.yaml", yamlData, 0644)
				if err != nil {
					fmt.Printf("Error while writing to file. %v", err)
				}

				s.Stop()
				fmt.Printf("Successfully authenticated, creds written to ~/.trakt.yaml\n")

				break
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.PersistentFlags().String("client-id", "", "")
	authCmd.PersistentFlags().String("client-secret", "", "")

	err := authCmd.MarkPersistentFlagRequired("client-id")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to mark client-id flag required")
	}
	err = authCmd.MarkPersistentFlagRequired("client-secret")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to mark client-secret flag required")
	}
}
