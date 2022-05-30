package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	jobprocessor "github.com/KompiTech/itsm-reporting-service/internal/domain/job/processor"
	"github.com/KompiTech/itsm-reporting-service/internal/domain/ref"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/api"
	converters "github.com/KompiTech/itsm-reporting-service/internal/http/rest/api/input_converters"
	"github.com/KompiTech/itsm-reporting-service/internal/mocks"
	"github.com/KompiTech/itsm-reporting-service/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateJobHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	t.Parallel()

	t.Run("with invalid payload", func(t *testing.T) {
		jobProcessor := new(mocks.JobProcessorMock)

		server := NewServer(Config{
			Addr:                    "service.url",
			Logger:                  logger,
			JobsProcessor:           jobProcessor,
			ExternalLocationAddress: "http://service.url",
		})

		payload := []byte(`{"type":"some nonsense"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/jobs", body)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		defer func() { _ = resp.Body.Close() }()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("could not read response: %v", err)
		}

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Status code")

		expectedJSON := `{"error":"type must be one of ['FE report only' 'SD report only' 'all']"}`
		assert.JSONEq(t, expectedJSON, string(b), "response does not match")
	})

	t.Run("when the job processor is busy", func(t *testing.T) {
		jobID := ref.UUID("38316161-3035-4864-ad30-6231392d3433")
		failingJob := job.Job{}

		jobsSvc := new(mocks.JobServiceMock)
		jobsSvc.On("CreateJob", api.CreateJobParams{Type: job.TypeAll}).
			Return(jobID, nil)

		jobsSvc.On("GetJob", jobID).Return(failingJob, nil)
		failingJob.FinalStatus = "Error: job is being processed, try it later"
		jobsSvc.On("UpdateJob", failingJob).Return(jobID, nil)

		jobProcessor := new(mocks.JobProcessorMock)
		jobProcessor.On("ProcessNewJob", jobID).Return(jobprocessor.ErrorBusy)

		server := NewServer(Config{
			Addr:                    "service.url",
			Logger:                  logger,
			JobsService:             jobsSvc,
			JobsProcessor:           jobProcessor,
			ExternalLocationAddress: "http://service.url",
		})

		payload := []byte(`{"type":"all"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/jobs", body)

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
		jobsSvc.On("CreateJob", api.CreateJobParams{Type: job.TypeAll}).
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

		payload := []byte(`{"type":"all"}`)

		body := bytes.NewReader(payload)
		req := httptest.NewRequest("POST", "/jobs", body)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Status code")
		expectedLocation := "http://service.url/jobs/38316161-3035-4864-ad30-6231392d3433"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"), "Location header")
	})
}

func TestGetJobHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	uuid := "cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
	retJob := job.Job{
		CreatedAt:                  "2022-03-14T00:10:00+01:00",
		ChannelsDownloadFinishedAt: "2022-03-14T00:12:00+01:00",
		FinalStatus:                "success",
		Type:                       job.TypeAll,
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
		"final_status":"success",
		"type":"all",
		"created_at":"2022-03-14T00:10:00+01:00",
		"channels_download_finished_at":"2022-03-14T00:12:00+01:00",
		"uuid":"cb2fe2a7-ab9f-4f6d-9fd6-c7c209403cf0"
	}`
	assert.JSONEq(t, expectedJSON, string(b), "response does not match")
}

func TestListJobHandler(t *testing.T) {
	logger, _ := testutils.NewTestLogger()
	defer func() { _ = logger.Sync() }()

	var list []job.Job
	job1 := job.Job{
		Type:      job.TypeFE,
		CreatedAt: "2022-03-12T11:47:22+01:00",
	}
	err := job1.SetUUID("0756952a-da33-4fe0-a883-9f899444c859")
	require.NoError(t, err)

	job2 := job.Job{
		Type:      job.TypeAll,
		CreatedAt: "2022-03-13T21:14:33+01:00",
	}
	err = job2.SetUUID("78202ce5-68aa-4fec-9a80-b863ac38bc06")
	require.NoError(t, err)

	job3 := job.Job{
		Type:      job.TypeSD,
		CreatedAt: "2022-03-14T00:10:00+01:00",
	}
	err = job3.SetUUID("f7b7fc74-e740-4c5f-a348-e8dc35b987ab")
	require.NoError(t, err)

	list = append(list, job1, job2, job3)

	jobsSvc := new(mocks.JobServiceMock)
	jobsSvc.On("ListJobs", mock.MatchedBy(
		func(paginationParams converters.PaginationParams) bool { return paginationParams.Page() == 0 }),
	).Return(list, nil)

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
			"created_at": "2022-03-12T11:47:22+01:00",
			"type": "FE report only"
		},		
		{
			"uuid": "78202ce5-68aa-4fec-9a80-b863ac38bc06",
			"created_at": "2022-03-13T21:14:33+01:00",
			"type": "all"
		},
		{
			"uuid": "f7b7fc74-e740-4c5f-a348-e8dc35b987ab",
			"created_at": "2022-03-14T00:10:00+01:00",
			"type": "SD report only"
		}
	]`

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type header")

	assert.JSONEq(t, expectedJSON, string(b), "response does not match")
}
