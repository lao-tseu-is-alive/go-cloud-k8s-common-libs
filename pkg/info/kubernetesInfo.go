package info

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
)

type K8sInfo struct {
	CurrentNamespace string `json:"current_namespace"`
	Version          string `json:"version"`
	Token            string `json:"token"`
	CaCert           string `json:"ca_cert"`
}

type ErrorConfig struct {
	Err error
	Msg string
}

// Error returns a string with an error and a specifics message
func (e *ErrorConfig) Error() string {
	return fmt.Sprintf("%s : %v", e.Msg, e.Err)
}

// GetKubernetesApiUrlFromEnv returns the k8s api url based on the content of standard env var :
//
//	KUBERNETES_SERVICE_HOST
//	KUBERNETES_SERVICE_PORT
//	in case the above ENV variables doesn't  exist the function returns an empty string and an error
func GetKubernetesApiUrlFromEnv() (string, error) {
	srvPort := 443
	k8sApiUrl := "https://"

	var err error
	val, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if !exist {
		return "", &ErrorConfig{
			Err: err,
			Msg: "ERROR: KUBERNETES_SERVICE_HOST ENV variable does not exist (not inside K8s ?).",
		}
	}
	k8sApiUrl = fmt.Sprintf("%s%s", k8sApiUrl, val)
	val, exist = os.LookupEnv("KUBERNETES_SERVICE_PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			return "", &ErrorConfig{
				Err: err,
				Msg: "ERROR: CONFIG ENV PORT should contain a valid integer.",
			}
		}
		if srvPort < 1 || srvPort > 65535 {
			return "", &ErrorConfig{
				Err: err,
				Msg: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
			}
		}
	}
	return fmt.Sprintf("%s:%d", k8sApiUrl, srvPort), nil
}

// GetKubernetesConnInfo returns a K8sInfo with various information retrieved from the current k8s api url
// K8sInfo.CurrentNamespace contains the current namespace of the running pod
// K8sInfo.Token contains the bearer authentication token for the running k8s instance in which this pods lives
// K8sInfo.CaCert contains the certificate of the running k8s instance in which this pods lives
func GetKubernetesConnInfo(l *slog.Logger, defaultReadTimeout time.Duration) (*K8sInfo, ErrorConfig) {
	const K8sServiceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
	K8sNamespacePath := fmt.Sprintf("%s/namespace", K8sServiceAccountPath)
	K8sTokenPath := fmt.Sprintf("%s/token", K8sServiceAccountPath)
	K8sCaCertPath := fmt.Sprintf("%s/ca.crt", K8sServiceAccountPath)

	info := K8sInfo{
		CurrentNamespace: "",
		Version:          "",
		Token:            "",
		CaCert:           "",
	}

	K8sNamespace, err := os.ReadFile(K8sNamespacePath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading namespace in " + K8sNamespacePath,
		}
	}
	info.CurrentNamespace = string(K8sNamespace)

	K8sToken, err := os.ReadFile(K8sTokenPath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading token in " + K8sTokenPath,
		}
	}
	info.Token = string(K8sToken)

	K8sCaCert, err := os.ReadFile(K8sCaCertPath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading Ca Cert in " + K8sCaCertPath,
		}
	}
	info.CaCert = string(K8sCaCert)

	k8sUrl, err := GetKubernetesApiUrlFromEnv()
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading GetKubernetesApiUrlFromEnv ",
		}
	}
	urlVersion := fmt.Sprintf("%s/openapi/v2", k8sUrl)
	res, err := gohttpclient.GetJsonFromUrlWithBearerAuth(urlVersion, info.Token, K8sCaCert, false, defaultReadTimeout, l)
	if err != nil {

		l.Info("GetKubernetesConnInfo error in GetJsonFromUrl", "url", urlVersion, "error", err)
		//return &info, ErrorConfig{
		//	Err: Err,
		//	Msg: fmt.Sprintf("GetKubernetesConnInfo: error doing GetJsonFromUrl(url:%s)", urlVersion),
		//}
	} else {
		l.Info("GetKubernetesConnInfo successfully returned from GetJsonFromUrl", "url", urlVersion)
		var myVersionRegex = regexp.MustCompile("{\"title\":\"(?P<title>.+)\",\"version\":\"(?P<version>.+)\"}")
		match := myVersionRegex.FindStringSubmatch(strings.TrimSpace(res[:150]))
		k8sVersionFields := make(map[string]string)
		for i, name := range myVersionRegex.SubexpNames() {
			if i != 0 && name != "" {
				k8sVersionFields[name] = match[i]
			}
		}
		info.Version = fmt.Sprintf("%s, %s", k8sVersionFields["title"], k8sVersionFields["version"])
	}

	return &info, ErrorConfig{
		Err: nil,
		Msg: "",
	}
}
