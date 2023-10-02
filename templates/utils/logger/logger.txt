package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.elastic.co/apm/v2"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var once sync.Once
var log zerolog.Logger

type Fields struct {
	Id       string    `json:"id"`
	RemoteIp string    `json:"remote_ip"`
	Host     string    `json:"host"`
	Method   string    `json:"method"`
	Header   string    `json:"header"`
	Request  string    `json:"request"`
	Response string    `json:"response"`
	Protocol string    `json:"protocol"`
	StartAt  time.Time `json:"start_at"`
	Path     string    `json:"path"`
	Error    error     `json:"error"`
}

func Log() zerolog.Logger {
	once.Do(func() {
		var isDebugMode bool
		isDebugMode, _ = strconv.ParseBool(os.Getenv("DEBUG_MODE"))
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if isDebugMode {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		var gitRevision string
		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			pwd, _ := os.Getwd()
			file = strings.ReplaceAll(file, pwd+"/", "")
			return file + ":" + strconv.Itoa(line)
		}

		log = zerolog.New(os.Stdout).
			With().
			Stack().
			Timestamp().
			Caller().
			Str("git_revision", gitRevision).
			Str("go_version", buildInfo.GoVersion).
			Logger()
	})

	return log
}

func LogOutboundRequest(ctx context.Context, path string, header []byte, request []byte, response []byte, statusCode int, fields Fields, errs []error) {
	defer catch()
	var isDebugMode bool
	isDebugMode, _ = strconv.ParseBool(os.Getenv("DEBUG_MODE"))

	var l = Log()
	var logs *zerolog.Event

	switch strconv.Itoa(statusCode)[0] {
	case '1', '2', '3':
		logs = l.Info()
	case '4', '5':
		l.Error()
		logs = l.Error()
	default:
		logs = l.Panic()
	}

	if len(errs) != 0 {
		logs.Err(errors.Wrap(errs[0], "Reason"))
	}

	if !isDebugMode {
		header = []byte("{}")
	}

	if len(request) != 0 {
		logs.RawJSON("request", request)
	}

	if json.Valid(response) {
		logs.RawJSON("response", response)
	} else {
		logs.Str("response", string(response))
	}
	tx := apm.TransactionFromContext(ctx)

	logs.
		Str("path", path).
		Str("trace_id", tx.TraceContext().Trace.String()).
		RawJSON("header", header).
		Int("status_code", statusCode).
		Str("id", fields.Id).
		Str("method", fields.Method).
		Str("protocol", fields.Protocol).
		Float64("latency", time.Since(fields.StartAt).Seconds()).
		Msg(fmt.Sprintf("%v %v callback %v", statusCode, fields.Method, path))
}

func LogMiddleware(path string,
	ctx context.Context,
	header []byte,
	queryParam []byte,
	request []byte,
	response []byte,
	statusCode int,
	fields Fields,
	err error,
) {
	defer catch()
	var l = Log()
	var logs *zerolog.Event

	switch strconv.Itoa(statusCode)[0] {
	case '1', '2', '3':
		logs = l.Info()
	case '4', '5':
		l.Error()
		logs = l.Error()
	default:
		logs = l.Panic()
	}

	if err != nil {
		logs.Err(errors.Wrap(err, "Reason"))
	}

	if len(queryParam) != 0 {
		logs.Str("query_param", string(queryParam))
	}

	if len(request) != 0 {
		dst := &bytes.Buffer{}
		if err := json.Compact(dst, request); err != nil {
			panic(err)
		}

		logs.RawJSON("request", dst.Bytes())
	}

	tx := apm.TransactionFromContext(ctx)

	logs.
		Str("path", path).
		Str("trace_id", tx.TraceContext().Trace.String()).
		RawJSON("header", header).
		RawJSON("response", response).
		Int("status_code", statusCode).
		Str("id", fields.Id).
		Str("remote_ip", fields.RemoteIp).
		Str("host", fields.Host).
		Str("method", fields.Method).
		Str("protocol", fields.Protocol).
		Float64("latency", time.Since(fields.StartAt).Seconds()).
		Msg(fmt.Sprintf("%v %v %v", statusCode, fields.Method, path))
}

func catch() {
	rvr := recover()
	if rvr != nil {
		var ok bool
		l := Log()

		err, ok := rvr.(error)
		if !ok {
			err = fmt.Errorf("%v", rvr)
		}

		l.Error().Err(err).Msg(err.Error())
	}
}
