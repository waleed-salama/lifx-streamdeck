package streamdeck

import (
	"github.com/stretchr/testify/require"
	"gitlab.com/wwsean08/lifx-streamdeck/mocks"
	"testing"
)

func TestRediscoverJob_Execute(t *testing.T) {
	testController := new(mocks.Controller)
	testController.On("DiscoverDevices").Return(nil)

	job := RediscoverJob{controller: testController}
	job.Execute()

	testController.AssertExpectations(t)
}

func TestRediscoverJob_Description(t *testing.T) {
	job := RediscoverJob{description: "test"}
	require.Equal(t, "test", job.description)

	job2 := RediscoverJob{}
	require.Zero(t, job2.Description())
}

func TestRediscoverJob_Key(t *testing.T) {
	job := RediscoverJob{}
	require.Equal(t, 42, job.Key())
}
