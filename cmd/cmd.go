package cmd

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"os/signal"
	"strconv"

	"os"
	"strings"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"

	"github.com/jothflee/honeygogo/backend"
	"github.com/jothflee/honeygogo/core"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var Fwd = ""
var FwdUser = ""
var FwdPw = ""
var FwdTLS = false
var FwdServer = ""
var ESIndex = "honeygogo"
var envPrefix = "HGG"
var LogLevelStr = "info"
var Port = "10025"
var Banner = "localhost"
var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "api",
	Long:  `api server`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
		err := initializeConfig(cmd)
		if err == nil {
			setLogLevel(LogLevelStr)
		}
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		out := core.StartSMTPServer(fmt.Sprintf(":%s", Port), Banner)
		be := selectBackend("elasticsearch", ESIndex)
		fwder := parseFwder()

		// catch sig
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, os.Interrupt)

		for {
			select {
			case in := <-out:
				if be != nil {
					be.OnMessage(in)
				}
				if fwder != nil {
					sendEmail(fwder, in)
				}
				log.Infof("%s", core.JSONstringify(in))
			case <-sigC:
				goto cleanup
			}

		}
	cleanup:
		be.Close()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}

func sendEmail(fwdTo *Forwarder, msg core.MessageMeta) {
	from := msg.To
	user := fwdTo.User
	password := fwdTo.Password
	to := fwdTo.Address
	tmp := strings.Split(fwdTo.Host, ":")
	smtpHost := tmp[0]
	smtpPort := 25

	if len(tmp) > 1 {
		i64, err := strconv.ParseInt(tmp[1], 10, 64)
		if err == nil {
			smtpPort = int(i64)
		}
	}

	d := gomail.NewDialer(smtpHost, smtpPort, user, password)
	if fwdTo.TLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	d.LocalName = msg.ToDomain

	mr := bytes.NewReader(msg.Data)

	closer, err := d.Dial()
	if err != nil {
		log.Fatal(err)
	}

	closer.Send(from, []string{to}, mr)
	err = closer.Close()
	if err != nil {
		log.Debug(err)
	}
}
func setLogLevel(level string) {
	switch level {
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	}
}
func selectBackend(backendType, name string) backend.Backend {
	switch backendType {
	case "elasticsearch":
		return backend.NewESBackend(name)
	}

	return nil
}

func init() {

	rootCmd.Flags().StringVarP(&ESIndex, "index", "i", "honeygogo", "the namespace/index/database name that will be used to store data in the backend (default: honeygogo)")
	rootCmd.Flags().StringVarP(&LogLevelStr, "log", "l", "info", "log level (trace, debug, warn, error, fatal")
	rootCmd.Flags().StringVarP(&Port, "port", "p", "10025", "the port to listen on (default: 10025)")
	rootCmd.Flags().StringVarP(&Banner, "banner", "b", "localhost", "the domain presented on the smtp banner (default: localhost)")
	rootCmd.Flags().StringVar(&Fwd, "fwd-addr", "", "the address to fwd all mail to (default: nowhere)")
	rootCmd.Flags().StringVar(&FwdUser, "fwd-user", "", "the user to login to the fwd mailserver (default: '')")
	rootCmd.Flags().StringVar(&FwdPw, "fwd-pw", "", "the password to login to the fwd mailserver (default: '')")
	rootCmd.Flags().StringVar(&FwdServer, "fwd-server", "", "the host (host:port) of the mailserver (default: '')")
	rootCmd.Flags().BoolVar(&FwdTLS, "fwd-tls", false, "the mailserver is using tls (default: false)")

}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	// v.SetConfigName(defaultConfigFilename)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	// v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
