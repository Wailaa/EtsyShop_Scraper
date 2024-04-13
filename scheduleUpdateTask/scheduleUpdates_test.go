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

type MockShopController struct {
	mock.Mock
}

func (m *MockShopController) UpdateSellingHistory(shop *models.Shop, task *models.TaskSchedule, shopRequest *models.ShopRequest) error {
	args := m.Called(shop, task, shopRequest)
	return args.Error(0)

}

func TestScheduleScrapUpdate_SchedulesCronJob(t *testing.T) {

	cronJob := &MockCronJob{}

	err := ScheduleScrapUpdate(cronJob)

	assert.Nil(t, err)

	assert.True(t, cronJob.AddFuncCalled)
	assert.True(t, cronJob.StartCalled)
	assert.Equal(t, "12 15 * * *", cronJob.AddFuncArg1)
}

func TestUpdateSoldItems_ShopParameterNil(t *testing.T) {

	shopController := &MockShopController{}

	queue := UpdateSoldItemsQueue{
		Shop: models.Shop{},
		Task: models.TaskSchedule{},
	}
	shopController.On("UpdateSellingHistory", mock.AnythingOfType("*models.Shop"), mock.AnythingOfType("*models.TaskSchedule"), mock.AnythingOfType("*models.ShopRequest")).Return(nil)

	UpdateSoldItems(queue, shopController)

	shopController.AssertCalled(t, "UpdateSellingHistory", mock.AnythingOfType("*models.Shop"), mock.AnythingOfType("*models.TaskSchedule"), mock.AnythingOfType("*models.ShopRequest"))

}
