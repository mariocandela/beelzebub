package MCP

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

type remoteAddrCtxKey struct{}

type MCPStrategy struct {
}

func (mcpStrategy *MCPStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	mcpServer := server.NewMCPServer(
		servConf.Description,
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	for _, toolConfig := range servConf.Tools {
		if toolConfig.Params == nil || len(toolConfig.Params) == 0 {
			log.Errorf("Tool %s has no parameters defined", toolConfig.Name)
			continue
		}

		opts := []mcp.ToolOption{
			mcp.WithDescription(toolConfig.Description),
		}

		for _, param := range toolConfig.Params {
			opts = append(opts,
				mcp.WithString(
					param.Name,
					mcp.Required(),
					mcp.Description(param.Description),
				),
			)
		}

		tool := mcp.NewTool(toolConfig.Name, opts...)

		mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			host, port, _ := net.SplitHostPort(ctx.Value(remoteAddrCtxKey{}).(string))

			tr.TraceEvent(tracer.Event{
				Msg:           "New MCP tool invocation",
				Protocol:      tracer.MCP.String(),
				Status:        tracer.Stateless.String(),
				RemoteAddr:    ctx.Value(remoteAddrCtxKey{}).(string),
				SourceIp:      host,
				SourcePort:    port,
				ID:            uuid.New().String(),
				Description:   servConf.Description,
				Command:       fmt.Sprintf("%s|%s", request.Params.Name, request.Params.Arguments),
				CommandOutput: toolConfig.Handler,
			})
			return mcp.NewToolResultText(toolConfig.Handler), nil
		})
	}

	go func() {
		httpServer := server.NewStreamableHTTPServer(
			mcpServer,
			server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				return context.WithValue(ctx, remoteAddrCtxKey{}, r.RemoteAddr)
			}),
		)
		if err := httpServer.Start(servConf.Address); err != nil {
			log.Errorf("Failed to start MCP server on %s: %v", servConf.Address, err)
			return
		}
	}()
	log.WithFields(log.Fields{
		"port":        servConf.Address,
		"description": servConf.Description,
	}).Infof("Init service %s", servConf.Protocol)
	return nil
}
