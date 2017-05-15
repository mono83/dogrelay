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
	Use:   "logstash-sentry",
	Short: "Runs logstash to sentry forwarder",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := v.All(
			v.WithMessage(v.StringNotWhitespace(ltsBind), "Empty bind address"),
			v.WithMessage(v.StringNotWhitespace(ltsDsn), "Empty sentry DSN"),
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
	Type    string `json:"type"`
	Name    string `json:"name"`
	Release string `json:"release"`
	Level   string `json:"log-level"`

	// Logger markers
	Marker1 string `json:"object"`
	Marker2 string `json:"marker"`

	// Messages
	HMessage string `json:"hmessage"`
	Message  string `json:"message"`
	Pattern  string `json:"pattern"`

	// Exceptions
	ExceptionMessage []string `json:"exception-message"`
	ExceptionClass   []string `json:"exception-class"`
	ExceptionTrace   []string `json:"exception-trace"`

	// Incoming extra
	Host      string `json:"host"`
	Instance  string `json:"instance"`
	RayID     string `json:"rayId"`
	SessionID string `json:"sessionId"`
}

func (i incomingLogstashPacket) getMessage() string {
	if len(i.HMessage) > 0 {
		return i.HMessage
	}

	return i.Message
}

func (i incomingLogstashPacket) getPattern() string {
	if len(i.Pattern) > 0 {
		return i.Pattern
	} else if len(i.Message) > 0 {
		return i.Message
	}

	return i.getMessage()
}

func (i incomingLogstashPacket) getSeverity() sentry.Severity {
	switch strings.ToLower(i.Level) {
	case "error":
		return sentry.ERROR
	case "alert", "critical", "emergency":
		return sentry.FATAL
	case "notice", "warning":
		return sentry.WARNING
	default:
		return sentry.INFO
	}
}

func (i incomingLogstashPacket) getMarker() string {
	if len(i.Marker1) > 0 {
		return i.Marker1
	} else if len(i.Marker2) > 0 {
		return i.Marker2
	}
	return "unknown"
}

func (i incomingLogstashPacket) getRayID() string {
	if len(i.RayID) > 0 {
		return i.RayID
	}

	return i.SessionID
}

func (i incomingLogstashPacket) getHost() string {
	if len(i.Instance) > 0 {
		return i.Instance
	}

	return i.Host
}

func (i incomingLogstashPacket) toSimple() sentry.SimplePacket {
	sim := sentry.SimplePacket{
		Message:   i.getMessage(),
		Level:     i.getSeverity(),
		Logger:    i.getMarker(),
		Release:   i.Release,
		Timestamp: sentry.Timestamp(time.Now()),
		Host:      i.getHost(),
		Extra:     map[string]string{},
	}

	// Generating fingerprint
	sim.Fingerprint = []string{i.getMarker(), i.getPattern(), string(i.getSeverity())}

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

	// Extra
	if r := i.getRayID(); len(r) > 0 {
		sim.Extra["rayId"] = r
	}

	return sim
}
