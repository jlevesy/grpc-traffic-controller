package gtc

import (
	"fmt"
	"path"
	"strconv"
	"strings"
)

func routeConfigName(namespace, name string) string {
	return path.Join(namespace, name, "routeconfig")
}

func vHostName(namespace, name string) string {
	return path.Join(namespace, name, "vhost")
}

func backendName(namespace, name string, routeID, backendID int) string {
	return path.Join(
		namespace,
		name,
		"route",
		strconv.Itoa(routeID),
		"backend",
		strconv.Itoa(backendID),
	)
}

// namespace/name/route/<route_id>/backend/<backend_id>
type parsedBackendName struct {
	Namespace    string
	ListenerName string
	RouteID      int
	BackendID    int
}

func (p *parsedBackendName) String() string {
	return backendName(p.Namespace, p.ListenerName, p.RouteID, p.BackendID)
}

func parseBackendName(resourceName string) (parsedBackendName, error) {
	sp := strings.Split(resourceName, "/")

	if len(sp) != 6 {
		return parsedBackendName{}, malformedResourceNameErr(resourceName)
	}

	routeID, err := strconv.Atoi(sp[3])
	if err != nil {
		return parsedBackendName{}, malformedResourceNameErr(resourceName)
	}

	backendID, err := strconv.Atoi(sp[5])
	if err != nil {
		return parsedBackendName{}, malformedResourceNameErr(resourceName)
	}

	return parsedBackendName{
		Namespace:    sp[0],
		ListenerName: sp[1],
		RouteID:      routeID,
		BackendID:    backendID,
	}, nil
}

type malformedResourceNameErr string

func (m malformedResourceNameErr) Error() string {
	return fmt.Sprintf("malformed resource name %q", string(m))
}
