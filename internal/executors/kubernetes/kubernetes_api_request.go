package kubernetes

import (
	"fmt"
	"os"
	"strings"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/cache"
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/http"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	API_VERSION           string = "apiVersion"
	NAMESPACE             string = "namespace"
	RESOURCE_TYPE         string = "resourceType"
	RESOURCE_NAME         string = "resourceName"
	PATH                  string = "path"
	QUERY                 string = "query"
	SHOULD_USE_LOCAL_DATA string = "shouldUseLocalData"
	KUBECONFIG            string = "kubeconfig"
	CERT_AUTHORITY_DATA   string = "certificateAuthorityData"
	TOKEN                 string = "token"

	RESPONSE_BODY   string = "body"
	RESPONSE_STATUS string = "status"

	USER                 string = "user"
	PASSWORD             string = "password"
	AUTHORIZATION_HEADER string = "authorizationHeader"
	SERVER               string = "server"
	URL                  string = "url"
	TRUSTED_CERTS        string = "trustedCerts"

	API_V1 string = "v1"
)

var API_V1_RESOURCE_TYPES = map[string]bool{"componentstatuses": true, "configmaps": true, "endpoints": true, "events": true, "limitranges": true, "namespaces": true, "persistentvolumeclaims": true, "pods": true, "podtemplates": true, "replicationcontrollers": true, "resourcequotas": true, "secrets": true, "serviceaccounts": true, "services": true, "nodes": true, "persistentvolumes": true}

type KubernetesApiRequestExecutor struct {
	executors.Executor
	httpExecutor http.HttpRequestExecutor
}

func (e *KubernetesApiRequestExecutor) Execute(ctx executors.ExecutorContext) *executors.ExecutorResult {
	config, err := buildConfig(ctx)

	if err != nil {
		return err
	}

	if shouldUseKubeconfig(ctx) && config == nil {
		return executors.NewExecutorResult(
			executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
			executors.ErrorString("Kubeconfig needed but could not be resolved"),
		)
	}

	input := ctx.GetInput()

	prepareBasicAuth(ctx, input, config)
	prepareToken(ctx, config, input)
	prepareUrl(ctx, config, input)
	prepareTrustedCerts(ctx, config, input)

	return buildOutput(e.httpExecutor.Execute(ctx))
}

func buildConfig(ctx executors.ExecutorContext) (*rest.Config, *executors.ExecutorResult) {
	useLocalData, err := ctx.GetBoolean(SHOULD_USE_LOCAL_DATA)
	// useLocalData = true

	if err != nil {
		return nil, nonRetriableExecutionResult(err)
	}

	if useLocalData {
		return getLocalKubeConfig(), nil
	} else {
		return buildProvidedKubeConfig(ctx)
	}
}

func getLocalKubeConfig() *rest.Config {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		os.Exit(1)
	}

	return config
}

func buildProvidedKubeConfig(ctx executors.ExecutorContext) (*rest.Config, *executors.ExecutorResult) {
	if ctx.GetString(KUBECONFIG) == "" {
		return nil, nil
	}

	kubeconfig, err := clientcmd.NewClientConfigFromBytes([]byte(ctx.GetString(KUBECONFIG)))
	if err != nil {
		return nil, nonRetriableExecutionResult(err)
	}

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, nonRetriableExecutionResult(err)
	}

	return config, nil
}

func shouldUseKubeconfig(ctx executors.ExecutorContext) bool {
	return ctx.GetString(SERVER) == "" ||
		(ctx.GetString(USER) == "" && ctx.GetString(TOKEN) == "") ||
		(ctx.GetString(PASSWORD) == "" && ctx.GetString(TOKEN) == "")
}

func prepareBasicAuth(ctx executors.ExecutorContext, input cache.MapCache[string, string], config *rest.Config) {
	user := ctx.GetString(USER)
	if user == "" && config != nil {
		user = config.Username
	}

	password := ctx.GetString(PASSWORD)
	if password == "" && config != nil {
		password = config.Password
	}

	input.Write(USER, user)
	input.Write(PASSWORD, password)
}

func prepareToken(ctx executors.ExecutorContext, config *rest.Config, input cache.MapCache[string, string]) {
	token := ctx.GetString(TOKEN)
	if token == "" && config != nil {
		token = config.BearerToken
	}

	if token != "" {
		auth, _ := http.NewBearerAuthorizationHeader(token).Generate()
		input.Write(AUTHORIZATION_HEADER, auth.GetValue())
	}
}

func prepareUrl(ctx executors.ExecutorContext, config *rest.Config, input cache.MapCache[string, string]) {
	server := ctx.GetString(SERVER)
	if server == "" && config != nil {
		server = config.Host
	}

	input.Write(URL, buildUrl(ctx, server))
}

func prepareTrustedCerts(ctx executors.ExecutorContext, config *rest.Config, input cache.MapCache[string, string]) {
	trustedCerts := ctx.GetString(CERT_AUTHORITY_DATA)
	if trustedCerts == "" && config != nil {
		trustedCerts = string(config.TLSClientConfig.CAData)
	}

	input.Write(TRUSTED_CERTS, trustedCerts)
}

func buildUrl(ctx executors.ExecutorContext, server string) string {
	var sb strings.Builder
	var apiPathName string
	apiVersion := ctx.GetString(API_VERSION)

	if API_V1_RESOURCE_TYPES[ctx.GetString(RESOURCE_TYPE)] && apiVersion == API_V1 {
		apiPathName = "api"
	} else {
		apiPathName = "apis"
	}

	sb.WriteString(fmt.Sprintf("%s/%s/%s", server, apiPathName, apiVersion))

	appendOptional(&sb, ctx, NAMESPACE, "/namespaces/%s")
	appendOptional(&sb, ctx, RESOURCE_TYPE, "/%s")
	appendOptional(&sb, ctx, RESOURCE_NAME, "/%s")
	appendOptional(&sb, ctx, PATH, "%s")
	appendOptional(&sb, ctx, QUERY, "%s")

	return sb.String()
}

func buildOutput(result *executors.ExecutorResult) *executors.ExecutorResult {
	output := make(map[string]any)
	output[RESPONSE_BODY] = result.Output[RESPONSE_BODY]
	output[RESPONSE_STATUS] = result.Output[RESPONSE_STATUS]

	return executors.NewExecutorResult(
		executors.Output(output),
		executors.Status(result.Status),
		executors.ErrorString(result.Error),
	)
}

func appendOptional(sb *strings.Builder, ctx executors.ExecutorContext, key string, valueFormat string) {
	value := ctx.GetString(key)
	if ctx.GetString(key) != "" {
		sb.Write([]byte(fmt.Sprintf(valueFormat, value)))
	}
}

func nonRetriableExecutionResult(err error) *executors.ExecutorResult {
	return executors.NewExecutorResult(
		executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_RETRYABLE),
		executors.Error(err),
	)
}
