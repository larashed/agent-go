package collectors

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/larashed/agent-go/monitoring/metrics"
)

func TestServiceListParser(t *testing.T) {
	str := `
	[ - ] apache2
	[ + ] nginx
	[ ? ] php-fpm7.3
	[ + ] php-fpm7.4
	[ 0 ] mysql
	`
	services := parseServiceList(str)

	assert.Equal(t, metrics.Service{
		Status: metrics.ServiceStatusStopped,
		Name:   "apache2",
	}, services[0])
	assert.Equal(t, metrics.Service{
		Status: metrics.ServiceStatusRunning,
		Name:   "nginx",
	}, services[1])
	assert.Equal(t, metrics.Service{
		Status: metrics.ServiceStatusUnknown,
		Name:   "php-fpm7.3",
	}, services[2])
	assert.Equal(t, metrics.Service{
		Status: metrics.ServiceStatusRunning,
		Name:   "php-fpm7.4",
	}, services[3])
	assert.Equal(t, 4, len(services))
}

