// build !ignore
package sender

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/larashed/agent-go/monitoring/buckets"
)

type ApiMock struct {
	wg                 *sync.WaitGroup
	recordChunkSize    int
	totalRecords       int
	appRecords         []int
	appRecordCallCount int
}

func (am *ApiMock) SendServerMetrics(data interface{}) error {
	return nil
}
func (am *ApiMock) SendApplicationRecords(data string) error {
	ints := strings.Split(data, "\n")
	for i := 0; i < len(ints); i++ {
		num, _ := strconv.Atoi(ints[i])
		am.appRecords = append(am.appRecords, num)
	}

	am.appRecordCallCount++

	if am.appRecordCallCount == am.totalRecords/am.recordChunkSize {
		am.wg.Done()
	}

	return nil
}

func (am *ApiMock) SendDeployment(data interface{}) error {
	return nil
}

func TestSender_SendApplicationMetrics(t *testing.T) {
	wg := sync.WaitGroup{}

	bucket := buckets.NewBucket()

	const recordChunkSize = 4
	const totalRecords = 100
	api := &ApiMock{
		wg:              &wg,
		recordChunkSize: recordChunkSize,
		totalRecords:    totalRecords,
	}

	sender := NewSender(
		api,
		bucket,
		recordChunkSize,
		1*time.Nanosecond,
		1*time.Nanosecond,
	)

	sender.SendApplicationMetrics()

	var expectedNumbers []int

	addRecords := func(from, to int) {
		defer wg.Done()

		for i := from; i < to; i++ {
			bucket.Add(strconv.Itoa(i))
			expectedNumbers = append(expectedNumbers, i)
		}
	}

	wg.Add(1)
	wg.Add(1)
	wg.Add(1)
	wg.Add(1)
	go addRecords(30, 60)
	go addRecords(60, 100)
	go addRecords(0, 30)

	wg.Wait()
	sender.Stop()

	sort.Ints(expectedNumbers)
	sort.Ints(api.appRecords)

	assert.Equal(t, totalRecords/recordChunkSize, api.appRecordCallCount)
	assert.Equal(t, expectedNumbers, api.appRecords)
	assert.Equal(t, len(expectedNumbers), len(api.appRecords))
}
