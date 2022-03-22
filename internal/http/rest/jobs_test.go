package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	jobprocessor "github.com/KompiTech/itsm-reporting-service/internal/domain/job/processor"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/KompiTech/itsm-reporting-service/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJobHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	t.Parallel()

	t.Run("when the job processor is busy", func(t *testing.T) {
		jobID := ref.UUID("38316161-3035-4864-ad30-6231392d3433")
		jobsSvc := new(mocks.JobServiceMock)
		jobsSvc.On("CreateJob").
			Return(jobID, nil)

		jobProcessor := new(mocks.JobProcessorMock)
		jobProcessor.On("ProcessNewJob", jobID).Return(jobprocessor.ErrorBusy)

		server := NewServer(Config{
			Addr:                    "service.url",
			Logger:                  logger,
			JobsService:             jobsSvc,
			JobsProcessor:           jobProcessor,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("POST", "/jobs", nil)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Status code")
		assert.Equal(t, "600", resp.Header.Get("Retry-After"), "Retry-After header")

		expectedJSON := `{"error":"job is being processed, try it later"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when the job processor is ready", func(t *testing.T) {
		jobID := ref.UUID("38316161-3035-4864-ad30-6231392d3433")
		jobsSvc := new(mocks.JobServiceMock)
		jobsSvc.On("CreateJob").
			Return(jobID, nil)

		jobProcessor := new(mocks.JobProcessorMock)
		jobProcessor.On("ProcessNewJob", jobID).Return(nil)

		server := NewServer(Config{
			Addr:                    "service.url",
			Logger:                  logger,
			JobsService:             jobsSvc,
			JobsProcessor:           jobProcessor,
			ExternalLocationAddress: "http://service.url",
		})

		req := httptest.NewRequest("POST", "/jobs", nil)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/jobs/38316161-3035-4864-ad30-6231392d3433"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})
}

func TestGetIncidentHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
	retJob := job.Job{
		CreatedAt:              "2022-03-14T00:10:00+01:00",
		ProcessingStartedAt:    "2022-03-14T00:12:00+01:00",
		ChannelsDownloadStatus: "success",
	}
	err := retJob.SetUUID(ref.UUID(uuid))
	require.NoError(t, err)

	jobsSvc := new(mocks.JobServiceMock)
	jobsSvc.On("GetJob", ref.UUID(uuid)).
		Return(retJob, nil)

	jobProcessor := new(mocks.JobProcessorMock)

	server := NewServer(Config{
		Addr:                    "service.url",
		Logger:                  logger,
		JobsService:             jobsSvc,
		JobsProcessor:           jobProcessor,
		ExternalLocationAddress: "http://service.url",
	})

	req := httptest.NewRequest("GET", "/jobs/"+uuid, nil)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	defer func() { _ = resp.Body.Close() }()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

	expectedJSON := `{
		"channels_download_status":"success",
		"created_at":"2022-03-14T00:10:00+01:00",
		"processing_started_at":"2022-03-14T00:12:00+01:00",
		"uuid":"cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
	}`
	assert.JSONEq(t, expectedJSON, string(b), "response does not match")
}

func TestListJobHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	var list []job.Job
	job1 := job.Job{
		CreatedAt: "2022-03-12T11:47:22+01:00",
	}
	err := job1.SetUUID("0756952a-da33-4fe0-a883-9f899444c859")
	require.NoError(t, err)

	job2 := job.Job{
		CreatedAt: "2022-03-13T21:14:33+01:00",
	}
	err = job2.SetUUID("78202ce5-68aa-4fec-9a80-b863ac38bc06")
	require.NoError(t, err)

	job3 := job.Job{
		CreatedAt: "2022-03-14T00:10:00+01:00",
	}
	err = job3.SetUUID("f7b7fc74-e740-4c5f-a348-e8dc35b987ab")
	require.NoError(t, err)

	list = append(list, job1, job2, job3)

	jobsSvc := new(mocks.JobServiceMock)
	jobsSvc.On("ListJobs").
		Return(list, nil)

	jobProcessor := new(mocks.JobProcessorMock)

	server := NewServer(Config{
		Addr:                    "service.url",
		Logger:                  logger,
		JobsService:             jobsSvc,
		JobsProcessor:           jobProcessor,
		ExternalLocationAddress: "http://service.url",
	})

	req := httptest.NewRequest("GET", "/jobs", nil)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	resp := w.Result()

	defer func() { _ = resp.Body.Close() }()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	expectedJSON := `[
		{
			"uuid": "0756952a-da33-4fe0-a883-9f899444c859",
			"created_at": "2022-03-12T11:47:22+01:00"
		},		
		{
			"uuid": "78202ce5-68aa-4fec-9a80-b863ac38bc06",
			"created_at": "2022-03-13T21:14:33+01:00"
		},
		{
			"uuid": "f7b7fc74-e740-4c5f-a348-e8dc35b987ab",
			"created_at": "2022-03-14T00:10:00+01:00"
		}
	]`

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

	assert.JSONEq(t, expectedJSON, string(b), "response does not match")
}
