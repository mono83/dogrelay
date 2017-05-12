package cmd

import (
	"encoding/json"
	"github.com/mono83/dogrelay/sentry"
	"github.com/mono83/dogrelay/udp"
	"github.com/mono83/slf/wd"
	v "github.com/mono83/validate"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var ltsBind, ltsDsn, ltsProject string

var logstashToSentryCmd = &cobra.Command{
	Use: "logstash-sentry",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := v.All(
			v.WithMessage(v.StringNotWhitespace(ltsBind), "Empty bind address"),
			v.WithMessage(v.StringNotWhitespace(ltsDsn), "Empty sentry DSN"),
			v.WithMessage(v.StringNotWhitespace(ltsProject), "Empty sentry project name"),
		); err != nil {
			return err
		}

		log := wd.NewLogger("logstash-sentry")
		// Building sentry client

		client, err := sentry.NewClient(ltsDsn)
		if err != nil {
			log.Error("Sentry client failed - :err", wd.ErrParam(err))
			return err
		}
		log.Info("Sentry client initialized")

		// Starting UDP listener
		err = udp.StartServer(ltsBind, 8*4096, func(bts []byte) {
			// Reading
			var in incomingLogstashPacket
			if err := json.Unmarshal(bts, &in); err != nil {
				log.Warning("Unable to parse incoming JSON - :err", wd.ErrParam(err))
			} else {
				in.SentryProject = ltsProject
				client.Send(in.toSimple())
			}
		})
		log.Info("UDP listener in logstash format established at :addr", wd.StringParam("addr", ltsBind))

		for {
			time.Sleep(time.Second)
		}
	},
}

func init() {
	logstashToSentryCmd.Flags().StringVar(&ltsBind, "bind", "", "Listening port and address, for example localhost:8080")
	logstashToSentryCmd.Flags().StringVar(&ltsDsn, "dsn", "", "Sentry DSN")
	logstashToSentryCmd.Flags().StringVar(&ltsProject, "project", "", "Sentry project name")
}

type incomingLogstashPacket struct {
	SentryProject    string   `json:"-"`
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	Release          string   `json:"release"`
	Level            string   `json:"log-level"`
	Marker1          string   `json:"object"`
	Marker2          string   `json:"marker"`
	Message          string   `json:"message"`
	Host             string   `json:"host"`
	ExceptionMessage []string `json:"exception-message"`
	ExceptionClass   []string `json:"exception-class"`
	ExceptionTrace   []string `json:"exception-trace"`
}

func (i incomingLogstashPacket) getMarker() string {
	if len(i.Marker1) > 0 {
		return i.Marker1
	} else if len(i.Marker2) > 0 {
		return i.Marker2
	}
	return "unknown"
}

func (i incomingLogstashPacket) toSimple() sentry.SimplePacket {
	sev := sentry.WARNING
	switch strings.ToLower(i.Level) {
	case "error":
		sev = sentry.ERROR
	case "alert", "critical", "emergency":
		sev = sentry.FATAL
	case "notice", "warning":
		sev = sentry.WARNING
	}

	sim := sentry.SimplePacket{
		Message:   i.Message,
		Release:   i.Release,
		Level:     sev,
		Timestamp: sentry.Timestamp(time.Now()),
		Logger:    i.getMarker(),
		Host:      i.Host,
	}

	// Generating fingerprint
	sim.Fingerprint = []string{sim.Logger, sim.Message, string(sim.Level)}

	if len(i.ExceptionClass) > 0 && len(i.ExceptionMessage) == len(i.ExceptionClass) {
		es := sentry.SimpleExceptions{}
		for j, cls := range i.ExceptionClass {
			e := sentry.SimpleException{
				Message: i.ExceptionMessage[j],
				Class:   cls,
			}
			if j == 0 && len(i.ExceptionTrace) > 0 {
				tr := sentry.SimpleStacktrace{}
				for _, l := range i.ExceptionTrace {
					tr.Frames = append(tr.Frames, sentry.SimpleFrame{Function: l})
				}
				e.Stacktrace = &tr
			}
			es.Values = append(es.Values, e)
		}
		sim.Exception = &es
	}

	return sim
}
