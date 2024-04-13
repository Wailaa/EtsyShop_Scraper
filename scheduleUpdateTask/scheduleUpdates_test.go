package scheduleUpdates

import (
	"EtsyScraper/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCronJob struct {
	AddFuncCalled bool
	AddFuncArg1   string
	StartCalled   bool
}

func (m *MockCronJob) AddFunc(spec string, cmd func()) {
	m.AddFuncCalled = true
	m.AddFuncArg1 = spec
}

func (m *MockCronJob) Start() {
	m.StartCalled = true
}

func TestScheduleScrapUpdate_SchedulesCronJob(t *testing.T) {

	cronJob := &MockCronJob{}

	err := ScheduleScrapUpdate(cronJob)

	assert.Nil(t, err)

	assert.True(t, cronJob.AddFuncCalled)
	assert.True(t, cronJob.StartCalled)
	assert.Equal(t, "12 15 * * *", cronJob.AddFuncArg1)
}

