package gtc

import (
	"fmt"
	"strings"
)

func resourcePrefix(namespace, name string) string {
	return "gtc" + "." + namespace + "." + name + "."
}

func routeConfigName(namespace, name string) string {
	return resourcePrefix(namespace, name) + "routeconfig"
}

func vHostName(namespace, name string) string {
	return resourcePrefix(namespace, name) + "vhost"
}

func clusterName(namespace, name, clusterName string) string {
	return resourcePrefix(namespace, name) + clusterName
}

type xdsResourceName struct {
	Namespace    string
	ListenerName string
	ResourceName string
}

func parseXDSResourceName(resourceName string) (xdsResourceName, error) {
	sp := strings.Split(resourceName, ".")

	if len(sp) != 4 {
		return xdsResourceName{}, malformedResourceNameErr(resourceName)
	}

	if sp[0] != "gtc" {
		return xdsResourceName{}, malformedResourceNameErr(resourceName)
	}

	return xdsResourceName{
		Namespace:    sp[1],
		ListenerName: sp[2],
		ResourceName: sp[3],
	}, nil
}

type malformedResourceNameErr string

func (m malformedResourceNameErr) Error() string {
	return fmt.Sprintf("malformed resource name %q", string(m))
}
