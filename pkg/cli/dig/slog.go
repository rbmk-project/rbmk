// SPDX-License-Identifier: GPL-3.0-or-later

package dig

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/miekg/dns"
)

// slogHandler handles logs possibly printing messages.
//
// Use [*Task.newSlogHandler] to construct.
type slogHandler struct {
	ch   slog.Handler
	task *Task
}

// newSlogHandler constructs a new [*slogHandler].
func (task *Task) newSlogHandler() *slogHandler {
	return &slogHandler{
		ch:   slog.NewJSONHandler(task.LogsWriter, &slog.HandlerOptions{}),
		task: task,
	}
}

var _ slog.Handler = &slogHandler{}

// Enabled implements [slog.Handler].
func (h *slogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements [slog.Handler].
func (h *slogHandler) Handle(ctx context.Context, rec slog.Record) error {
	// Intercept and print the raw query and the raw response
	switch rec.Message {
	case "dnsQuery":
		rec.Attrs(func(attr slog.Attr) bool {
			if attr.Key != "dnsRawQuery" {
				return true
			}
			if rawQuery, ok := attr.Value.Any().([]byte); ok {
				h.printRawQuery(rawQuery)
			}
			return false
		})

	case "dnsResponse":
		rec.Attrs(func(attr slog.Attr) bool {
			if attr.Key != "dnsRawResponse" {
				return true
			}
			if rawResp, ok := attr.Value.Any().([]byte); ok {
				h.printRawResponse(rawResp)
			}
			return false
		})
	}

	// Defer to the underlying logger
	return h.ch.Handle(ctx, rec)
}

// WithAttrs implements [slog.Handler].
func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.ch = h.ch.WithAttrs(attrs)
	return h
}

// WithGroup implements [slog.Handler].
func (h *slogHandler) WithGroup(name string) slog.Handler {
	h.ch = h.ch.WithGroup(name)
	return h
}

func (h *slogHandler) printRawQuery(rawQuery []byte) {
	query := &dns.Msg{}
	if err := query.Unpack(rawQuery); err != nil {
		return
	}
	fmt.Fprintf(h.task.QueryWriter, ";; Query:\n%s\n", query.String())
}

func (h *slogHandler) printRawResponse(rawResp []byte) {
	resp := &dns.Msg{}
	if err := resp.Unpack(rawResp); err != nil {
		return
	}
	fmt.Fprintf(h.task.ResponseWriter, "\n;; Response:\n%s\n\n", resp.String())
	fmt.Fprintf(h.task.ShortWriter, "%s", h.formatResponseShort(resp))
}

func (h *slogHandler) formatResponseShort(response *dns.Msg) string {
	var builder strings.Builder
	for _, ans := range response.Answer {
		switch ans := ans.(type) {
		case *dns.A:
			fmt.Fprintf(&builder, "%s\n", ans.A.String())

		case *dns.AAAA:
			fmt.Fprintf(&builder, "%s\n", ans.AAAA.String())

		case *dns.CNAME:
			if !h.task.ShortIP {
				fmt.Fprintf(&builder, "%s\n", ans.Target)
			}

		case *dns.HTTPS:
			if !h.task.ShortIP {
				value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
				fmt.Fprintf(&builder, "%s\n", value)
			}

		case *dns.MX:
			if !h.task.ShortIP {
				value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
				fmt.Fprintf(&builder, "%s\n", value)
			}

		case *dns.NS:
			if !h.task.ShortIP {
				value := strings.TrimPrefix(ans.String(), ans.Hdr.String())
				fmt.Fprintf(&builder, "%s\n", value)
			}

		default:
			// TODO(bassosimone): implement the other answer types
		}
	}
	return builder.String()
}
