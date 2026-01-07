package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/OpenNHP/opennhp/nhp/log"
)

var (
	taApiPrefix    = fmt.Sprintf("%s/ta", serviceApiPrefix)
	bufferedTaMap  = make(map[string]*TrustedApplication)
	bufferedTaLock sync.Mutex
)

type TAFunctionParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type TAFunction struct {
	Method      string            `json:"method"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Params      []TAFunctionParam `json:"params"`
}

type TrustedApplication struct {
	Id        string
	Functions []TAFunction
	Ctx       context.Context
	Cancel    context.CancelFunc
	Client    *client.Client
}

func NewTrustApplication(tadId string, language string, entry string) (*TrustedApplication, error) {
	bufferedTaLock.Lock()
	defer bufferedTaLock.Unlock()
	if _, exists := bufferedTaMap[tadId]; exists {
		return bufferedTaMap[tadId], nil
	}

	ta := &TrustedApplication{
		Id:        tadId,
		Functions: []TAFunction{},
	}

	ctx, cancel := context.WithCancel(context.Background())

	var c *client.Client
	var err error

	stdioTransport := transport.NewStdio(entry, nil)
	c = client.NewClient(stdioTransport)

	if err := c.Start(ctx); err != nil {
		log.Error("Failed to start trusted application: %v", err)
		cancel()
		return nil, err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "Trusted Application Executor",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		log.Error("Failed to initialize: %v", err)
		cancel()
		return nil, err
	}

	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		log.Error("Failed to list functions which are supported in trusted application: %v", err)
		cancel()
		return nil, err
	} else {
		for _, tool := range toolsResult.Tools {
			taFunc := TAFunction{
				Method:      "POST",
				Name:        fmt.Sprintf("%s/%s/%s", taApiPrefix, ta.Id, tool.Name),
				Description: tool.Description,
				Params: []TAFunctionParam{
					{
						Name:        "doId",
						Description: "identifier of the data object",
						Type:        "string",
					},
				},
			}

			schema := tool.InputSchema
			for name, propSchema := range schema.Properties {
				if name == "path" { // path is injected by nhp agent
					continue
				}
				prop, _ := propSchema.(map[string]any)
				taFuncParam := TAFunctionParam{
					Name:        name,
					Description: prop["description"].(string),
					Type:        prop["type"].(string),
				}
				taFunc.Params = append(taFunc.Params, taFuncParam)
			}
			ta.Functions = append(ta.Functions, taFunc)
		}
	}

	ta.Ctx = ctx
	ta.Cancel = cancel
	ta.Client = c

	if _, exists := bufferedTaMap[tadId]; !exists {
		bufferedTaMap[tadId] = ta
	}

	return ta, nil
}

func GetTrustedApplication(trustedAppUuid string) (*TrustedApplication, error) {
	bufferedTaLock.Lock()
	defer bufferedTaLock.Unlock()

	if ta, exists := bufferedTaMap[trustedAppUuid]; exists {
		return ta, nil
	} else {
		return nil, fmt.Errorf("TrustedApplication not found, please register first")
	}
}

func (ta *TrustedApplication) GetSupportedFunctions() []TAFunction {
	return ta.Functions
}

func (ta *TrustedApplication) CallFunction(function string, params map[string]any) (string, error) {
	callRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      function,
			Arguments: params,
		},
	}

	callResponse, err := ta.Client.CallTool(ta.Ctx, callRequest)
	if err != nil {
		return "", err
	}

	// check the type of content
	switch firstContent := callResponse.Content[0].(type) {
	case mcp.TextContent:
		return firstContent.Text, nil
	default:
		return "", fmt.Errorf("unexpected content type: %T", callResponse.Content[0])
	}
}
